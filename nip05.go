package main

import (
	"encoding/json"
	"net/http"
	"path"
	"time"
)

type Cache struct {
	resp     Response
	lastSeen time.Time
}

var cache = *new(map[string]Cache)

type Response struct {
	Names  map[string]string   `json:"names"`
	Relays map[string][]string `json:"relays"`
}

func NIP05Handler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	if name == "" {
		http.Error(w, "Missing query parameter 'name'", http.StatusBadRequest)
		return
	}

	c, exist := cache[name]
	if exist {
		c.lastSeen = time.Now()

		_ = json.NewEncoder(w).Encode(c.resp)

		return
	}

	doc := loadNIP05()
	if doc == nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	resp := Response{
		Names:  make(map[string]string, 1),
		Relays: make(map[string][]string, 1),
	}

	pubKey, ok := doc.Names[name]
	if !ok {
		http.Error(w, "Can't find this name", http.StatusNotFound)
		return
	}

	resp.Names[name] = pubKey

	relays, ok := doc.Relays[pubKey]
	if ok {
		resp.Relays[pubKey] = relays
	}

	cache[name] = Cache{
		lastSeen: time.Now(),
		resp:     resp,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func loadNIP05() *Response {
	resp := new(Response)
	data, err := ReadFile(path.Join(config.WorkingDirectory, "/nip05.json"))
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(data, resp); err != nil {
		return nil
	}

	return resp
}

func setNIP05(pubkey, name string) error {
	resp := new(Response)
	data, err := ReadFile(path.Join(config.WorkingDirectory, "/nip05.json"))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, resp); err != nil {
		return err
	}

	resp.Names[name] = pubkey

	data, err = json.Marshal(resp)
	if err != nil {
		return err
	}

	if err := WriteFile(path.Join(config.WorkingDirectory, "/nip05.json"), data); err != nil {
		return err
	}

	return nil
}

func unSetNIP05(name string) error {
	resp := new(Response)
	data, err := ReadFile(path.Join(config.WorkingDirectory, "/nip05.json"))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, resp); err != nil {
		return err
	}

	delete(resp.Names, name)

	data, err = json.Marshal(resp)
	if err != nil {
		return err
	}

	if err := WriteFile(path.Join(config.WorkingDirectory, "/nip05.json"), data); err != nil {
		return err
	}

	return nil
}

func checkCache() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		for name, c := range cache {
			if time.Since(c.lastSeen) >= time.Hour {
				delete(cache, name)
			}
		}
	}
}
