package main

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

func ReceiveReport(_ context.Context, reportEvt *nostr.Event) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyReported := management.ModerationEvents[reportEvt.ID]
	if alreadyReported {
		return fmt.Errorf("already received this report: %s", reportEvt.ID)
	}

	management.ModerationEvents[reportEvt.ID] = reportEvt.Content

	return nil
}
