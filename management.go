package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"path"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip86"
)

var management *Management = &Management{
	Mutex: *new(sync.Mutex),
}

type Management struct {
	AllowedPubkeys   []string `json:"allowed_keys"`
	BannedPubkeys    []string `json:"banned_keys"`
	DisallowedKins   []int    `json:"disallowed_kinds"`
	AllowedKinds     []int    `json:"allowed_kinds"`
	BlockedIPs       []string `json:"blocked_ips"`
	BannedEvents     []string `json:"banned_events"`
	ModerationEvents []string `json:"moderation_events"`

	sync.Mutex
}

func AllowPubkey(_ context.Context, pubkey, _ string) error {
	management.Lock()
	defer management.Unlock()

	for i, p := range management.BannedPubkeys {
		if p == pubkey {
			management.BannedPubkeys[i] = ""
		}
	}

	management.AllowedPubkeys = append(management.AllowedPubkeys, pubkey)

	UpdateManagement()

	return nil
}

func BanPubkey(_ context.Context, pubkey, _ string) error {
	management.Lock()
	defer management.Unlock()

	for _, q := range relay.QueryEvents {
		ech, err := q(context.Background(), nostr.Filter{
			Authors: []string{pubkey},
		})
		if err != nil {
			return err
		}

		for evt := range ech {
			for _, dl := range relay.DeleteEvent {
				if err := dl(context.Background(), evt); err != nil {
					return err
				}
			}
		}
	}

	management.BannedPubkeys = append(management.BannedPubkeys, pubkey)

	UpdateManagement()

	return nil
}

func AllowKind(_ context.Context, kind int) error {
	management.Lock()
	defer management.Unlock()

	for i, k := range management.DisallowedKins {
		if k == kind {
			management.DisallowedKins[i] = -1
		}
	}

	management.AllowedKinds = append(management.AllowedKinds, kind)

	UpdateManagement()

	return nil
}

func DisallowKind(_ context.Context, kind int) error {
	management.Lock()
	defer management.Unlock()

	management.DisallowedKins = append(management.DisallowedKins, kind)

	UpdateManagement()

	return nil
}

func BlockIP(_ context.Context, ip net.IP, reason string) error {
	management.Lock()
	defer management.Unlock()

	management.BlockedIPs = append(management.BlockedIPs, ip.String())

	UpdateManagement()

	return nil
}

func UnblockIP(_ context.Context, ip net.IP, reason string) error {
	management.Lock()
	defer management.Unlock()

	for i, bannedIP := range management.BlockedIPs {
		if bannedIP == ip.String() {
			management.BlockedIPs[i] = ""
		}
	}

	UpdateManagement()

	return nil
}

func BanEvent(_ context.Context, id string, reason string) error {
	management.Lock()
	defer management.Unlock()

	for _, q := range relay.QueryEvents {
		ech, err := q(context.Background(), nostr.Filter{
			IDs: []string{id},
		})
		if err != nil {
			return err
		}

		for evt := range ech {
			for _, dl := range relay.DeleteEvent {
				if err := dl(context.Background(), evt); err != nil {
					return err
				}
			}
		}
	}

	management.BannedEvents = append(management.BannedEvents, id)

	UpdateManagement()

	return nil
}

func ListAllowedKinds(_ context.Context) ([]int, error) {
	management.Lock()
	defer management.Unlock()

	return management.AllowedKinds, nil
}

func ListAllowedPubKeys(_ context.Context) ([]nip86.PubKeyReason, error) {
	management.Lock()
	defer management.Unlock()

	// todo:: support reason.
	res := []nip86.PubKeyReason{}
	for _, p := range management.AllowedPubkeys {
		res = append(res, nip86.PubKeyReason{
			PubKey: p,
		})
	}

	return res, nil
}

func ListBannedEvents(_ context.Context) ([]nip86.IDReason, error) {
	management.Lock()
	defer management.Unlock()

	// todo:: support reason.
	res := []nip86.IDReason{}
	for _, id := range management.BannedEvents {
		res = append(res, nip86.IDReason{
			ID: id,
		})
	}

	return res, nil
}

func ListBannedPubKeys(_ context.Context) ([]nip86.PubKeyReason, error) {
	management.Lock()
	defer management.Unlock()

	// todo:: support reason.
	res := []nip86.PubKeyReason{}
	for _, p := range management.BannedPubkeys {
		res = append(res, nip86.PubKeyReason{
			PubKey: p,
		})
	}

	return res, nil
}

func ListBlockedIPs(ctx context.Context) ([]nip86.IPReason, error) {
	management.Lock()
	defer management.Unlock()

	// todo:: support reason.
	res := []nip86.IPReason{}
	for _, ip := range management.BlockedIPs {
		res = append(res, nip86.IPReason{
			IP: ip,
		})
	}

	return res, nil
}

func ListEventsNeedingModeration(_ context.Context) ([]nip86.IDReason, error) {
	management.Lock()
	defer management.Unlock()

	// todo:: support reason.
	res := []nip86.IDReason{}
	for _, id := range management.ModerationEvents {
		res = append(res, nip86.IDReason{
			ID: id,
		})
	}

	return res, nil
}

func LoadManagement() {
	if !PathExists(path.Join(config.WorkingDirectory, "/management.json")) {
		data, err := json.Marshal(new(Management))
		if err != nil {
			log.Fatalf("can't make management.json: %s", err.Error())
		}

		if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
			log.Fatalf("can't make management.json: %s", err.Error())
		}
	}

	data, err := ReadFile(path.Join(config.WorkingDirectory, "/management.json"))
	if err != nil {
		log.Fatalf("can't read management.json: %s", err.Error())
	}

	if err := json.Unmarshal(data, management); err != nil {
		log.Fatalf("can't read management.json: %s", err.Error())
	}
}

func UpdateManagement() {
	data, err := json.Marshal(management)
	if err != nil {
		log.Fatalf("can't update management.json: %s", err.Error())
	}

	if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
		log.Fatalf("can't update management.json: %s", err.Error())
	}
}
