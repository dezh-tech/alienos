package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"path"
	"slices"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip86"
)

var management *Management = &Management{
	Mutex: *new(sync.Mutex),
}

type Management struct {
	AllowedPubkeys   map[string]string `json:"allowed_keys"`
	BannedPubkeys    map[string]string `json:"banned_keys"`
	DisallowedKins   []int             `json:"disallowed_kinds"`
	AllowedKinds     []int             `json:"allowed_kinds"`
	BlockedIPs       map[string]string `json:"blocked_ips"`
	BannedEvents     map[string]string `json:"banned_events"`
	ModerationEvents map[string]string `json:"moderation_events"`

	sync.Mutex
}

func AllowPubkey(_ context.Context, pubkey, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyAllowed := management.AllowedPubkeys[pubkey]
	if alreadyAllowed {
		return fmt.Errorf("pubkey %s is already allowed", pubkey)
	}

	delete(management.BannedPubkeys, pubkey)

	management.AllowedPubkeys[pubkey] = reason

	go sendNotification(fmt.Sprintf("Pubkey %s is now allowed on relay %s\nReason: %s",
		pubkey, config.RelayURL, reason))

	UpdateManagement()

	return nil
}

func BanPubkey(_ context.Context, pubkey, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyBanned := management.BannedPubkeys[pubkey]
	if alreadyBanned {
		return fmt.Errorf("pubkey %s is already banned", pubkey)
	}

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

	delete(management.AllowedPubkeys, pubkey)

	management.BannedPubkeys[pubkey] = reason

	go sendNotification(fmt.Sprintf("Pubkey %s is now banned on relay %s\nReason: %s",
		pubkey, config.RelayURL, reason))

	UpdateManagement()

	return nil
}

func AllowKind(_ context.Context, kind int) error {
	management.Lock()
	defer management.Unlock()

	if slices.Contains(management.AllowedKinds, kind) {
		return fmt.Errorf("kind %d is already allowed", kind)
	}

	management.DisallowedKins = slices.DeleteFunc(management.DisallowedKins, func(k int) bool {
		return k == kind
	})

	management.AllowedKinds = append(management.AllowedKinds, kind)

	go sendNotification(fmt.Sprintf("Kind %d is now allowed on relay %s",
		kind, config.RelayURL))

	UpdateManagement()

	return nil
}

func DisallowKind(_ context.Context, kind int) error {
	management.Lock()
	defer management.Unlock()

	if slices.Contains(management.DisallowedKins, kind) {
		return fmt.Errorf("kind %d is already disallowed", kind)
	}

	management.AllowedKinds = slices.DeleteFunc(management.AllowedKinds, func(k int) bool {
		return k == kind
	})

	management.DisallowedKins = append(management.DisallowedKins, kind)

	go sendNotification(fmt.Sprintf("Kind %d is now disallowed on relay %s",
		kind, config.RelayURL))

	UpdateManagement()

	return nil
}

func BlockIP(_ context.Context, ip net.IP, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyBlocked := management.BlockedIPs[ip.String()]
	if alreadyBlocked {
		return fmt.Errorf("ip %s is already blocked", ip.String())
	}

	management.BlockedIPs[ip.String()] = reason

	go sendNotification(fmt.Sprintf("IP %s is now blocked on relay %s\nReason: %s",
		ip.String(), config.RelayURL, reason))

	UpdateManagement()

	return nil
}

func UnblockIP(_ context.Context, ip net.IP, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, blocked := management.BlockedIPs[ip.String()]
	if !blocked {
		return fmt.Errorf("ip %s is not blocked", ip.String())
	}

	delete(management.BlockedIPs, ip.String())

	go sendNotification(fmt.Sprintf("IP %s is now unblocked on relay %s\nReason: %s",
		ip.String(), config.RelayURL, reason))

	UpdateManagement()

	return nil
}

func BanEvent(_ context.Context, id string, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyBanned := management.BannedEvents[id]
	if alreadyBanned {
		return fmt.Errorf("event %s is already banned", id)
	}

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

	management.BannedEvents[id] = reason

	go sendNotification(fmt.Sprintf("Event %s is now blocked on relay %s\nReason: %s",
		id, config.RelayURL, reason))

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

	res := []nip86.PubKeyReason{}
	for pubkey, reason := range management.AllowedPubkeys {
		res = append(res, nip86.PubKeyReason{
			PubKey: pubkey,
			Reason: reason,
		})
	}

	return res, nil
}

func ListBannedEvents(_ context.Context) ([]nip86.IDReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.IDReason{}
	for id, reason := range management.BannedEvents {
		res = append(res, nip86.IDReason{
			ID:     id,
			Reason: reason,
		})
	}

	return res, nil
}

func ListBannedPubKeys(_ context.Context) ([]nip86.PubKeyReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.PubKeyReason{}
	for pubkey, reason := range management.BannedPubkeys {
		res = append(res, nip86.PubKeyReason{
			PubKey: pubkey,
			Reason: reason,
		})
	}

	return res, nil
}

func ListBlockedIPs(ctx context.Context) ([]nip86.IPReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.IPReason{}
	for ip, reason := range management.BlockedIPs {
		res = append(res, nip86.IPReason{
			IP:     ip,
			Reason: reason,
		})
	}

	return res, nil
}

func ListEventsNeedingModeration(_ context.Context) ([]nip86.IDReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.IDReason{}
	for id, reason := range management.ModerationEvents {
		res = append(res, nip86.IDReason{
			ID:     id,
			Reason: reason,
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
