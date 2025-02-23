package main

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"slices"
	"syscall"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/bluge"
	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/blossom"
	"github.com/kehiy/blobstore"
	"github.com/kehiy/blobstore/disk"
	"github.com/kehiy/blobstore/minio"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/keyer"
	"github.com/nbd-wtf/go-nostr/nip86"
)

var (
	relay      *khatru.Relay
	config     Config
	plainKeyer nostr.Keyer
	simplePool *nostr.SimplePool

	//go:embed static/index.html
	landingTempl []byte
)

func main() {
	log.SetPrefix("alienos ")
	log.Printf("Running %s\n", StringVersion())
	LoadConfig()

	relay = khatru.NewRelay()

	relay.Info.Name = config.RelayName
	relay.Info.Description = config.RelayDescription
	relay.Info.Icon = config.RelayIcon
	relay.Info.Contact = config.RelayContact
	relay.Info.PubKey = config.RelayPublicKey
	relay.Info.URL = config.RelayURL
	relay.Info.Version = StringVersion()
	relay.Info.Software = "https://github.com/dezh-tech/alienos"

	relay.Info.SupportedNIPs = []any{1, 9, 11, 17, 40, 42, 50, 56, 59, 70, 86}

	badgerDB := badger.BadgerBackend{
		Path: path.Join(config.WorkingDirectory, "/db"),
	}

	if err := badgerDB.Init(); err != nil {
		log.Fatalf("can't setup db: %s", err.Error())
	}

	blugeDB := bluge.BlugeBackend{
		Path:          path.Join(config.WorkingDirectory, "/search_db"),
		RawEventStore: &badgerDB,
	}

	if err := blugeDB.Init(); err != nil {
		log.Fatalf("can't setup db: %s", err.Error())
	}

	relay.StoreEvent = append(relay.StoreEvent, badgerDB.SaveEvent, blugeDB.SaveEvent, StoreEvent)
	relay.QueryEvents = append(relay.QueryEvents, blugeDB.QueryEvents, badgerDB.QueryEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, badgerDB.DeleteEvent, blugeDB.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, badgerDB.ReplaceEvent, blugeDB.ReplaceEvent)
	relay.CountEvents = append(relay.CountEvents, badgerDB.CountEvents)
	relay.CountEventsHLL = append(relay.CountEventsHLL, badgerDB.CountEventsHLL)

	relay.RejectFilter = append(relay.RejectFilter, RejectFilter)
	relay.RejectEvent = append(relay.RejectEvent, RejectEvent)

	bl := blossom.New(relay, fmt.Sprintf("http://%s:%s", config.RelayBind, config.RelayPort))

	bl.Store = blossom.EventStoreBlobIndexWrapper{Store: &badgerDB, ServiceURL: bl.ServiceURL}

	if !PathExists(path.Join(config.WorkingDirectory, "/blossom")) {
		if err := Mkdir(path.Join(config.WorkingDirectory, "/blossom")); err != nil {
			log.Fatalf("can't initialize blossom directory: %s", err.Error())
		}
	}

	var blobStorage blobstore.Store

	blobStorage = disk.New(path.Join(config.WorkingDirectory, "/blossom"))

	if config.S3ForBlossom {
		blobStorage = minio.New(config.S3Endpoint, config.S3AccessKeyID,
			config.S3SecretKey, true, config.S3BlossomBucket, "")

		if err := blobStorage.Init(context.Background()); err != nil {
			log.Fatalf("can't init s3 for blossom: %s\n", err.Error())
		}
	}

	bl.StoreBlob = append(bl.StoreBlob, blobStorage.Store)
	bl.LoadBlob = append(bl.LoadBlob, blobStorage.Load)
	bl.DeleteBlob = append(bl.DeleteBlob, blobStorage.Delete)
	bl.ReceiveReport = append(bl.ReceiveReport, ReceiveReport)

	LoadManagement()

	relay.ManagementAPI.AllowPubKey = AllowPubkey
	relay.ManagementAPI.BanPubKey = BanPubkey
	relay.ManagementAPI.AllowKind = AllowKind
	relay.ManagementAPI.DisallowKind = DisallowKind
	relay.ManagementAPI.BlockIP = BlockIP
	relay.ManagementAPI.UnblockIP = UnblockIP
	relay.ManagementAPI.BanEvent = BanEvent
	relay.ManagementAPI.ListAllowedKinds = ListAllowedKinds
	relay.ManagementAPI.ListAllowedPubKeys = ListAllowedPubKeys
	relay.ManagementAPI.ListBannedEvents = ListBannedEvents
	relay.ManagementAPI.ListBannedPubKeys = ListBannedPubKeys
	relay.ManagementAPI.ListBlockedIPs = ListBlockedIPs
	relay.ManagementAPI.ListEventsNeedingModeration = ListEventsNeedingModeration
	relay.ManagementAPI.RejectAPICall = append(relay.ManagementAPI.RejectAPICall,
		func(ctx context.Context, mp nip86.MethodParams) (reject bool, msg string) {
			auth := khatru.GetAuthed(ctx)
			if !slices.Contains(config.Admins, auth) {
				return true, "your are not an admin"
			}

			return false, ""
		})

	mux := relay.Router()

	mux.HandleFunc("GET /{$}", StaticViewHandler)
	mux.HandleFunc("/.well-known/nostr.json", NIP05Handler)

	if config.BackupEnabled {
		go backupWorker()
	}

	simplePool = nostr.NewSimplePool(context.Background())
	pKeyer, err := keyer.NewPlainKeySigner(config.RelaySelf)
	if err != nil {
		log.Fatalf("can't create keyer: %s", err.Error())
	}

	plainKeyer = pKeyer

	log.Printf("Serving on ws://%s\n", config.RelayBind+config.RelayPort)
	go http.ListenAndServe(config.RelayBind+config.RelayPort, relay)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-sigChan

	log.Print("Received signal: Initiating graceful shutdown", "signal", sig.String())
	badgerDB.Close()
	blugeDB.Close()
	relay.Shutdown(context.Background())
}

func StaticViewHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	t := template.New("webpage")
	t, err := t.Parse(string(landingTempl))
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)

		return
	}

	err = t.Execute(w, relay.Info)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)

		return
	}
}
