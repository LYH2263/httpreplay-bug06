package studio

import "net/http"

func RegisterRoutes(mux *http.ServeMux, api http.Handler) {
	mux.Handle("/api/", api)
}
