package studio

import (
	"encoding/json"
	"net/http"

	"github.com/LYH2263/go-httpreplay"
)

type API struct {
	Cassette *httpreplay.Cassette
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/stats":
		writeJSON(w, a.Cassette.Stats())
	case "/api/interactions":
		writeJSON(w, a.Cassette.Interactions())
	case "/api/snapshot":
		writeJSON(w, a.Cassette.Snapshot())
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
