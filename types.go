package httpreplay

import (
	"net/http"
	"time"
)

type Interaction struct {
	ID              string
	Method          string
	URL             string
	RequestHeaders  map[string][]string
	RequestBody     []byte
	Status          int
	ResponseHeaders map[string][]string
	ResponseBody    []byte
	RecordedAt      time.Time
}

type Stats struct {
	Recorded  int
	Replayed  int
	Misses    int
	Unmatched int
	Pending   int
}

type Snapshot struct {
	Name         string
	Interactions []Interaction
	Stats        Stats
}

func cloneHeader(h http.Header) map[string][]string {
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
