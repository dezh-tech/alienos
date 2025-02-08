package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

func ReceiveReport(_ context.Context, reportEvt *nostr.Event) error {
	management.Lock()
	defer management.Unlock()

	management.ModerationEvents = append(management.ModerationEvents, reportEvt.ID)
	return nil
}
