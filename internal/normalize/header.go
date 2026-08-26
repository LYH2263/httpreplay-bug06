package normalize

import "net/http"

func StripVolatileHeaders(h http.Header, names []string) {
	for _, n := range names {
		h.Del(n)
	}
}

func CloneHeader(h http.Header) map[string][]string {
	if h == nil {
		return nil
	}
	out := make(map[string][]string, len(h))
	for k, vv := range h {
		cp := make([]string, len(vv))
		copy(cp, vv)
		out[k] = cp
	}
	return out
}
