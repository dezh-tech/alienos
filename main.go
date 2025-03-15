package main

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"slices"
	"strconv"
	"syscall"
	"time"

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
	relay           *khatru.Relay
	config          Config
	plainKeyer      nostr.Keyer
	simplePool      *nostr.SimplePool
	liveConnections int
	startTime       time.Time

	//go:embed static/index.html
	landingTempl []byte
)

func main() {
	Info("Running", "version", StringVersion())

	InitGlobalLogger()

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

	relay.Info.AddSupportedNIPs([]int{1, 9, 11, 17, 40, 42, 50, 56, 59, 70, 86})

	relay.OnConnect = append(relay.OnConnect, func(_ context.Context) {
		liveConnections++
	})

	relay.OnDisconnect = append(relay.OnDisconnect, func(_ context.Context) {
		liveConnections--
	})

	badgerDB := badger.BadgerBackend{
		Path: path.Join(config.WorkingDirectory, "/db"),
	}

	if err := badgerDB.Init(); err != nil {
		Fatal("can't setup db", "err", err.Error())
	}

	blugeDB := bluge.BlugeBackend{
		Path:          path.Join(config.WorkingDirectory, "/search_db"),
		RawEventStore: &badgerDB,
	}

	if err := blugeDB.Init(); err != nil {
		Fatal("can't setup db", "err", err.Error())
	}

	relay.StoreEvent = append(relay.StoreEvent, badgerDB.SaveEvent, blugeDB.SaveEvent, StoreEvent)
	relay.QueryEvents = append(relay.QueryEvents, blugeDB.QueryEvents, badgerDB.QueryEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, badgerDB.DeleteEvent, blugeDB.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, badgerDB.ReplaceEvent, blugeDB.ReplaceEvent)
	relay.CountEvents = append(relay.CountEvents, badgerDB.CountEvents)
	relay.CountEventsHLL = append(relay.CountEventsHLL, badgerDB.CountEventsHLL)

	relay.RejectFilter = append(relay.RejectFilter, RejectFilter)
	relay.RejectEvent = append(relay.RejectEvent, RejectEvent)

	bl := blossom.New(relay, fmt.Sprintf("http://%s:%d", config.RelayBind, config.RelayPort))

	bl.Store = blossom.EventStoreBlobIndexWrapper{Store: &badgerDB, ServiceURL: bl.ServiceURL}

	if !PathExists(path.Join(config.WorkingDirectory, "/blossom")) {
		if err := Mkdir(path.Join(config.WorkingDirectory, "/blossom")); err != nil {
			Fatal("can't initialize blossom directory", "err", err.Error())
		}
	}

	var blobStorage blobstore.Store

	blobStorage = disk.New(path.Join(config.WorkingDirectory, "/blossom"))

	if config.S3ForBlossom {
		blobStorage = minio.New(config.S3Endpoint, config.S3AccessKeyID,
			config.S3SecretKey, true, config.S3BlossomBucket, "")

		if err := blobStorage.Init(context.Background()); err != nil {
			Fatal("can't init s3 for blossom", "err", err.Error())
		}
	}

	bl.StoreBlob = append(bl.StoreBlob, blobStorage.Store)
	bl.LoadBlob = append(bl.LoadBlob, blobStorage.Load)
	bl.DeleteBlob = append(bl.DeleteBlob, blobStorage.Delete)
	bl.ReceiveReport = append(bl.ReceiveReport, ReceiveReport)

	LoadManagement()

	for _, admin := range config.Admins {
		_, isAdmin := management.Admins[admin]
		if isAdmin {
			continue
		}

		management.Admins[admin] = []string{"*"}

		UpdateManagement()
	}

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
	relay.ManagementAPI.Stats = Stats
	relay.ManagementAPI.GrantAdmin = GrantAdmin
	relay.ManagementAPI.RevokeAdmin = RevokeAdmin
	relay.ManagementAPI.Generic = Generic
	relay.ManagementAPI.RejectAPICall = append(relay.ManagementAPI.RejectAPICall,
		func(ctx context.Context, mp nip86.MethodParams) (reject bool, msg string) {
			auth := khatru.GetAuthed(ctx)
			methods, isAdmin := management.Admins[auth]
			if !isAdmin {
				return true, "your are not an admin"
			}

			if !slices.Contains(methods, "*") {
				if !slices.Contains(methods, mp.MethodName()) {
					return true, "you don't have access to this method"
				}
			}

			return false, ""
		})

	mux := relay.Router()

	mux.HandleFunc("GET /{$}", StaticViewHandler)

	mux.HandleFunc("/.well-known/nostr.json", NIP05Handler)
	go checkCache()

	if config.BackupEnabled {
		go backupWorker()
	}

	simplePool = nostr.NewSimplePool(context.Background())
	pKeyer, err := keyer.NewPlainKeySigner(config.RelaySelf)
	if err != nil {
		Fatal("can't create keyer", "err", err.Error())
	}

	plainKeyer = pKeyer

	startTime = time.Now()

	Info("Serving", "address", net.JoinHostPort(config.RelayBind, strconv.Itoa(config.RelayPort)))
	if err := relay.Start(config.RelayBind, config.RelayPort); err != nil {
		Error("can't start the server", "err", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-sigChan

	Info("Received signal: Initiating graceful shutdown", "signal", sig.String())
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
