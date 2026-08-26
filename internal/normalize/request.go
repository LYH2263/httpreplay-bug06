package normalize

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/LYH2263/go-httpreplay/internal/volatile"
)

type NormalizedRequest struct {
	Method string
	URL    string
	Body   []byte
}

func NormalizeRequest(req *http.Request, body []byte, cfg Config) NormalizedRequest {
	h := req.Header.Clone()
	StripVolatileHeaders(h, cfg.VolatileHeaders)
	u := URL(req.URL)
	if cfg.Rules.IgnoreQuery {
		u = stripQuery(u)
	}
	if cfg.Rules.IgnoreHost {
		u = stripHost(u)
	}
	return NormalizedRequest{
		Method: strings.ToUpper(req.Method),
		URL:    u,
		Body:   NormalizeBody(body, cfg),
	}
}

func URL(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.Scheme + "://" + u.Host + u.Path
}

func stripQuery(raw string) string {
	if i := strings.Index(raw, "?"); i >= 0 {
		return raw[:i]
	}
	return raw
}

func stripHost(raw string) string {
	if i := strings.Index(raw, "://"); i >= 0 {
		rest := raw[i+3:]
		if j := strings.Index(rest, "/"); j >= 0 {
			return rest[j:]
		}
	}
	return raw
}

type Config struct {
	VolatileHeaders []string
	Rules           Rules
}

type Rules struct {
	IgnoreQuery bool
	IgnoreHost  bool
	StripAuth   bool
}

func DefaultConfig() Config {
	return Config{
		VolatileHeaders: volatile.DefaultList(),
		Rules:           Rules{IgnoreQuery: false, IgnoreHost: false, StripAuth: true},
	}
}
