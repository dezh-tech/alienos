package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/nbd-wtf/go-nostr"
)

func ReceiveReport(_ context.Context, reportEvt *nostr.Event) error {
	management.Lock()
	defer management.Unlock()

	if slices.Contains(management.ModerationEvents, reportEvt.ID) {
		return fmt.Errorf("already received this report: %s", reportEvt.ID)
	}

	management.ModerationEvents = append(management.ModerationEvents, reportEvt.ID)
	return nil
}
