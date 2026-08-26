package httpreplay

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type staticTripper struct {
	status int
	body   string
}

func (s staticTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: s.status,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

func TestBug06_ReplayStatsRace(t *testing.T) {
	c, _ := OpenCassette(Options{Name: "demo"})
	for i := 0; i < 20; i++ {
		_ = c.Append(Interaction{
			ID:     fmt.Sprintf("id-%d", i),
			Method: http.MethodGet,
			URL:    fmt.Sprintf("http://example.com/%d", i),
			Status: 200,
		})
	}
	var wg sync.WaitGroup
	for n := 0; n < 8; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			p := NewPlayer(c, Options{})
			for i := 0; i < 20; i++ {
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://example.com/%d", i), nil)
				_, _ = p.RoundTrip(req)
				runtime.Gosched()
			}
		}(n)
	}
	wg.Wait()
}
