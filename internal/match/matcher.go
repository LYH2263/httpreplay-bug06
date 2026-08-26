package match

import (
	"bytes"

	"github.com/LYH2263/go-httpreplay/internal/normalize"
)

type Matcher interface {
	Match(req normalize.NormalizedRequest, cand Candidate, matchBody bool) bool
}

type Candidate struct {
	Method string
	URL    string
	Body   []byte
}

type DefaultMatcher struct {
	Rules Rules
}

func (m DefaultMatcher) Match(req normalize.NormalizedRequest, cand Candidate, matchBody bool) bool {
	if req.Method != cand.Method {
		return false
	}
	if req.URL != cand.URL {
		return false
	}
	if matchBody && !bytes.Equal(req.Body, cand.Body) {
		return false
	}
	return true
}
