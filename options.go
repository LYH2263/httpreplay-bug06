package httpreplay

import (
	"net/http"
	"time"

	"github.com/LYH2263/go-httpreplay/internal/clock"
	"github.com/LYH2263/go-httpreplay/internal/match"
	"github.com/LYH2263/go-httpreplay/internal/normalize"
)

type Options struct {
	Name          string
	StorePath     string
	Clock         clock.Clock
	Matcher       match.Matcher
	Normalize     normalize.Config
	RealTransport http.RoundTripper
	MatchBody     bool
	RecordTTL     time.Duration
}

func (o Options) withDefaults() Options {
	if o.Clock == nil {
		o.Clock = clock.System{}
	}
	if o.Matcher == nil {
		o.Matcher = match.DefaultMatcher{}
	}
	if o.Normalize.VolatileHeaders == nil {
		o.Normalize = normalize.DefaultConfig()
	}
	if o.RealTransport == nil {
		o.RealTransport = http.DefaultTransport
	}
	if o.RecordTTL <= 0 {
		o.RecordTTL = 24 * time.Hour
	}
	return o
}
