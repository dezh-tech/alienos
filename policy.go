package main

import (
	"context"
	"slices"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func RejectEvent(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	management.Lock()
	defer management.Unlock()

	if slices.Contains(management.BannedPubkeys, event.PubKey) {
		return true, "blocked: you are banned"
	}

	if config.WhiteListedPubkey {
		if !slices.Contains(management.AllowedPubkeys, event.PubKey) {
			return true, "restricted: you are not allowed"
		}
	}

	if slices.Contains(management.DisallowedKins, event.Kind) {
		return true, "blocked: kind not allowed"
	}

	if config.WhiteListedKind {
		if !slices.Contains(management.AllowedKinds, event.Kind) {
			return true, "restricted: kind not allowed"
		}
	}

	if slices.Contains(management.BannedEvents, event.ID) {
		return true, "blocked: event is banned"
	}

	if slices.Contains(management.BlockedIPs, khatru.GetIP(ctx)) {
		return true, "blocked: this IP is blocked"
	}

	return false, ""
}

func StoreEvent(ctx context.Context, event *nostr.Event) error {
	management.Lock()
	defer management.Unlock()

	if event.Kind == nostr.KindReporting {
		for _, t := range event.Tags {
			if t.Key() == "e" && t.Value() != "" {
				if len(t.Value()) == 64 {
					management.ModerationEvents = append(management.ModerationEvents, t.Value())
				}
			}
		}
	}

	UpdateManagement()

	return nil
}

func RejectUpload(ctx context.Context, auth *nostr.Event, size int, ext string) (bool, string, int) {
	management.Lock()
	defer management.Unlock()

	if slices.Contains(management.BannedPubkeys, auth.PubKey) {
		return true, "blocked: you are banned", 403
	}

	if config.WhiteListedPubkey {
		if !slices.Contains(management.AllowedPubkeys, auth.PubKey) {
			return true, "restricted: you are not allowed", 403
		}
	}

	if slices.Contains(management.BlockedIPs, khatru.GetIP(ctx)) {
		return true, "blocked: this IP is blocked", 403
	}

	return false, "", 200
}

func RejectFilter(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
	if !slices.Contains(filter.Kinds, nostr.KindGiftWrap) {
		return false, ""
	}

	auth := khatru.GetAuthed(ctx)
	if auth == "" {
		return true, "auth-required: you are reading DMs"
	}

	if len(filter.Tags) != 1 {
		return true, "error: you can read your DMs only"
	}

	if !(len(filter.Tags["#p"]) >= 2) {
		return true, "error: invalid p tag"
	}

	if filter.Tags["#p"][1] != auth {
		return true, "restricted: you can read people DMs"
	}

	return false, ""
}
