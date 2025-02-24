package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nbd-wtf/go-nostr/nip19"
)

func Mkdir(path string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("could not create directory %s", path)
	}

	return nil
}

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func WriteFile(filename string, data []byte) error {
	if err := Mkdir(filepath.Dir(filename)); err != nil {
		return err
	}
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write to %s: %w", filename, err)
	}

	return nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}

func HexPubkeyToMention(pubkey string) string {
	npub, err := nip19.EncodePublicKey(pubkey)
	if err != nil {
		return ""
	}

	return "nostr:" + npub
}

func HexEventIDToMention(id string) string {
	npub, err := nip19.EncodeNote(id)
	if err != nil {
		return ""
	}

	return "nostr:" + npub
}
