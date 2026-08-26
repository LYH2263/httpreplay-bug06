package httpreplay

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/LYH2263/go-httpreplay/internal/match"
	"github.com/LYH2263/go-httpreplay/internal/normalize"
)

type Player struct {
	cassette  *Cassette
	matcher   match.Matcher
	norm      normalize.Config
	matchBody bool
	mu        sync.Mutex
	used      map[string]bool
}

func NewPlayer(c *Cassette, opts Options) *Player {
	opts = opts.withDefaults()
	return &Player{
		cassette:  c,
		matcher:   opts.Matcher,
		norm:      opts.Normalize,
		matchBody: opts.MatchBody,
		used:      make(map[string]bool),
	}
}

func (p *Player) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	normReq := normalize.NormalizeRequest(req, body, p.norm)
	it, ok, err := p.findMatch(req.Context(), normReq, body)
	if err != nil {
		return nil, err
	}
	if !ok {
		p.cassette.bumpMiss()
		return nil, ErrNoMatch
	}
	p.cassette.bumpReplayed()
	return buildResponse(it), nil
}

func (p *Player) findMatch(ctx context.Context, req normalize.NormalizedRequest, body []byte) (Interaction, bool, error) {
	items := p.cassette.Interactions()
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, it := range items {
		if err := ctx.Err(); err != nil {
			return Interaction{}, false, err
		}
		if p.used[it.ID] {
			continue
		}
		candidate := match.Candidate{
			Method: it.Method,
			URL:    it.URL,
			Body:   it.RequestBody,
		}
		if !p.matcher.Match(req, candidate, p.matchBody) {
			continue
		}
		p.used[it.ID] = true
		return it, true, nil
	}
	p.cassette.bumpUnmatched()
	return Interaction{}, false, nil
}

func buildResponse(it Interaction) *http.Response {
	h := make(http.Header)
	for k, vv := range it.ResponseHeaders {
		for _, v := range vv {
			h.Add(k, v)
		}
	}
	return &http.Response{
		StatusCode: it.Status,
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(it.ResponseBody)),
	}
}

func (p *Player) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.used = make(map[string]bool)
}

func (p *Player) ReplayOnce(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := req.Clone(ctx)
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		out.Body = io.NopCloser(bytes.NewReader(b))
		req.Body = io.NopCloser(bytes.NewReader(b))
	}
	return p.RoundTrip(out)
}
