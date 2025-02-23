package main

import (
	"encoding/json"
	"net/http"
	"path"
)

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
