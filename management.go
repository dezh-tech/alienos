package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path"
	"slices"
	"sync"
	"time"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip86"
)

var management *Management = &Management{
	Mutex: *new(sync.Mutex),
}

type Management struct {
	AllowedPubkeys   map[string]string   `json:"allowed_keys"`
	BannedPubkeys    map[string]string   `json:"banned_keys"`
	DisallowedKins   []int               `json:"disallowed_kinds"`
	AllowedKinds     []int               `json:"allowed_kinds"`
	BlockedIPs       map[string]string   `json:"blocked_ips"`
	BannedEvents     map[string]string   `json:"banned_events"`
	ModerationEvents map[string]string   `json:"moderation_events"`
	Admins           map[string][]string `json:"admins"`

	sync.Mutex
}

type RelayStats struct {
	LiveConnections int     `json:"num_connections"`
	Uptime          float64 `json:"uptime"`
	// TotalBlobs   int `json:"total_blobs"` // TODO:::
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
		HexPubkeyToMention(pubkey), config.RelayURL, reason))

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
		HexPubkeyToMention(pubkey), config.RelayURL, reason))

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
		HexEventIDToMention(id), config.RelayURL, reason))

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

func Stats(_ context.Context) (nip86.Response, error) {
	return nip86.Response{
		Result: RelayStats{
			LiveConnections: liveConnections,
			Uptime:          time.Since(startTime).Seconds(),
		},
	}, nil
}

func GrantAdmin(ctx context.Context, pubkey string, methods []string) error {
	management.Lock()
	defer management.Unlock()

	if len(methods) == 0 {
		return errors.New("Methods can't be 0")
	}

	management.Admins[pubkey] = methods

	caller := khatru.GetAuthed(ctx)
	go sendNotification(fmt.Sprintf("New admin %s granted by %s\nMethods: %v",
		HexPubkeyToMention(pubkey), HexPubkeyToMention(caller), methods))

	UpdateManagement()

	return nil
}

func RevokeAdmin(ctx context.Context, pubkey string, methods []string) error {
	management.Lock()
	defer management.Unlock()

	_, isAdmin := management.Admins[pubkey]
	if !isAdmin {
		return fmt.Errorf("Pubkey %s is already not in admins list", pubkey)
	}

	management.Admins[pubkey] = slices.DeleteFunc(management.Admins[pubkey], func(m string) bool {
		return slices.Contains(methods, m)
	})

	deleted := false

	allowedMethods, _ := management.Admins[pubkey]
	if len(allowedMethods) == 0 {
		delete(management.Admins, pubkey)
		deleted = true
	}

	caller := khatru.GetAuthed(ctx)
	go sendNotification(fmt.Sprintf("Admin %s revoked by %s\nMethods: %v\nDeleted: %v",
		HexPubkeyToMention(pubkey), HexPubkeyToMention(caller), methods, deleted))

	UpdateManagement()

	return nil
}

func Generic(ctx context.Context, request nip86.Request) (nip86.Response, error) {
	switch request.Method {
	case "setnip5":
		if len(request.Params) != 2 {
			return nip86.Response{}, fmt.Errorf("invalid number of params for '%s'", request.Method)
		}

		pk, ok := request.Params[0].(string)
		if !ok || !nostr.IsValidPublicKey(pk) {
			return nip86.Response{}, fmt.Errorf("invalid pubkey param for '%s'", request.Method)
		}

		name, ok := request.Params[1].(string)
		if !ok {
			return nip86.Response{}, fmt.Errorf("invalid pubkey param for '%s'", request.Method)
		}

		if err := setNIP05(pk, name); err != nil {
			return nip86.Response{}, nil
		}

		go sendNotification(fmt.Sprintf("New NIP-05 has been set.\nName: %s\nPubkey: %s", name,
			HexPubkeyToMention(pk)))

		return nip86.Response{
			Result: "successful",
		}, nil

	case "unsetnip5":
		if len(request.Params) != 1 {
			return nip86.Response{}, fmt.Errorf("invalid number of params for '%s'", request.Method)
		}

		name, ok := request.Params[0].(string)
		if !ok {
			return nip86.Response{}, fmt.Errorf("invalid pubkey param for '%s'", request.Method)
		}

		if err := unSetNIP05(name); err != nil {
			return nip86.Response{}, nil
		}

		go sendNotification(fmt.Sprintf("NIP-05 has been unset.\nName: %s", name))

		return nip86.Response{
			Result: "successful",
		}, nil

	}

	return nip86.Response{}, fmt.Errorf("unknown method %s", request.Method)
}

func LoadManagement() {
	if !PathExists(path.Join(config.WorkingDirectory, "/management.json")) {
		data, err := json.Marshal(Management{
			AllowedPubkeys:   make(map[string]string),
			BannedPubkeys:    make(map[string]string),
			DisallowedKins:   make([]int, 0),
			AllowedKinds:     make([]int, 0),
			BlockedIPs:       make(map[string]string),
			BannedEvents:     make(map[string]string),
			ModerationEvents: make(map[string]string),
			Admins:           make(map[string][]string),
		})
		if err != nil {
			Fatal("can't make management.json", "err", err.Error())
		}

		if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
			Fatal("can't make management.json", "err", err.Error())
		}
	}

	data, err := ReadFile(path.Join(config.WorkingDirectory, "/management.json"))
	if err != nil {
		Fatal("can't read management.json", "err", err.Error())
	}

	if err := json.Unmarshal(data, management); err != nil {
		Fatal("can't read management.json", "err", err.Error())
	}
}

func UpdateManagement() {
	data, err := json.Marshal(management)
	if err != nil {
		Fatal("can't update management.json", "err", err.Error())
	}

	if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
		Fatal("can't update management.json", "err", err.Error())
	}
}
