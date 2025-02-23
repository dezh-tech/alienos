package main

import (
	"context"
	"log"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip17"
)

func sendNotification(text string) {
	for _, pubkey := range config.Admins {
		dmRelays := nip17.GetDMRelays(context.Background(), pubkey, simplePool, []string{
			"wss://nos.lol", "wss://jellyfish.land",
			"wss://nos.lol", "wss://relay.primal.net", "wss://relay.damus.io", "wss://relay.0xchat.com",
		})

		if err := nip17.PublishMessage(context.Background(), text, nostr.Tags{}, simplePool,
			dmRelays, dmRelays, plainKeyer, pubkey, nil); err != nil {
			log.Printf("can't send system notification: %s\n", err.Error())
		}
	}
}
