package main

import (
	"context"
	"net/http"
	"slices"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func RejectEvent(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	management.Lock()
	defer management.Unlock()

	_, banned := management.BannedPubkeys[event.PubKey]
	if banned {
		return true, "blocked: you are banned"
	}

	if config.WhiteListedPubkey {
		_, allowed := management.AllowedPubkeys[event.PubKey]
		if !allowed {
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

	_, eventBanned := management.BannedEvents[event.ID]
	if eventBanned {
		return true, "blocked: event is banned"
	}

	_, blocked := management.BlockedIPs[khatru.GetIP(ctx)]
	if blocked {
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
					management.ModerationEvents[t.Value()] = event.Content
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

	_, banned := management.BannedPubkeys[auth.PubKey]
	if banned {
		return true, "blocked: you are banned", http.StatusForbidden
	}

	if config.WhiteListedPubkey {
		_, allowed := management.AllowedPubkeys[auth.PubKey]
		if !allowed {
			return true, "restricted: you are not allowed", http.StatusForbidden
		}
	}

	_, blocked := management.BlockedIPs[khatru.GetIP(ctx)]
	if blocked {
		return true, "blocked: this IP is blocked", http.StatusForbidden
	}

	return false, "", http.StatusOK
}

// todo: can we handle it better?
func RejectFilter(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
	if !slices.Contains(filter.Kinds, nostr.KindGiftWrap) {
		return false, ""
	}

	auth := khatru.GetAuthed(ctx)
	if auth == "" {
		return true, "auth-required: you are reading gift-wrapped events"
	}

	if len(filter.Tags) != 1 {
		return true, "error: you can read your gift-wrapped events only"
	}

	if !(len(filter.Tags["#p"]) >= 2) {
		return true, "error: invalid p tag"
	}

	if filter.Tags["#p"][1] != auth {
		return true, "restricted: you can read people DMs"
	}

	return false, ""
}
