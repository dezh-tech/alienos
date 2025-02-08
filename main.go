package main

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/bluge"
	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/blossom"
	"github.com/kehiy/blobstore/disk"
)

var (
	relay  *khatru.Relay
	config Config
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

	relay.RejectFilter = append(relay.RejectFilter, RejectFilter)

	relay.RejectEvent = append(relay.RejectEvent, RejectEvent)

	bl := blossom.New(relay, fmt.Sprintf("http://%s:%s", config.RelayBind, config.RelayPort))

	bl.Store = blossom.EventStoreBlobIndexWrapper{Store: &badgerDB, ServiceURL: bl.ServiceURL}

	if !PathExists(path.Join(config.WorkingDirectory, "/blossom")) {
		if err := Mkdir(path.Join(config.WorkingDirectory, "/blossom")); err != nil {
			log.Fatalf("can't initialize blossom directory: %s", err.Error())
		}
	}

	blobStorage := disk.New(path.Join(config.WorkingDirectory, "/blossom"))

	bl.StoreBlob = append(bl.StoreBlob, blobStorage.Store)
	bl.LoadBlob = append(bl.LoadBlob, blobStorage.Load)
	bl.DeleteBlob = append(bl.DeleteBlob, blobStorage.Delete)
	bl.ReceiveReport = append(bl.ReceiveReport, ReceiveReport)

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

	mux := relay.Router()

	mux.HandleFunc("/.well-known/nostr.json", NIP05Handler)

	log.Printf("Serving on ws://%s\n", config.RelayBind+config.RelayPort)
	http.ListenAndServe(config.RelayBind+config.RelayPort, relay)
}
