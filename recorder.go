package httpreplay

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/LYH2263/go-httpreplay/internal/clone"
	"github.com/LYH2263/go-httpreplay/internal/normalize"
)

type Recorder struct {
	cassette  *Cassette
	transport http.RoundTripper
	normalize normalize.Config
	clock     clockInterface
}

type clockInterface interface {
	Now() time.Time
}

func NewRecorder(c *Cassette, opts Options) *Recorder {
	opts = opts.withDefaults()
	return &Recorder{
		cassette:  c,
		transport: opts.RealTransport,
		normalize: opts.Normalize,
		clock:     opts.Clock,
	}
}

func (r *Recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	body, err := readBody(req.Body)
	if err != nil {
		return nil, err
	}
	resp, err := r.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	respBody, err := readBody(resp.Body)
	if err != nil {
		return nil, err
	}
	it := Interaction{
		Method:          req.Method,
		URL:             normalize.URL(req.URL),
		RequestHeaders:  cloneHeader(req.Header),
		RequestBody:     clone.Bytes(body),
		Status:          resp.StatusCode,
		ResponseHeaders: cloneHeader(resp.Header),
		ResponseBody:    clone.Bytes(respBody),
		RecordedAt:      r.clock.Now(),
	}
	if err := r.cassette.Append(it); err != nil {
		return nil, err
	}
	out := cloneResponse(resp, respBody)
	return out, nil
}

func readBody(r io.ReadCloser) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	defer r.Close()
	return io.ReadAll(r)
}

func cloneResponse(resp *http.Response, body []byte) *http.Response {
	out := *resp
	out.Body = io.NopCloser(bytes.NewReader(body))
	out.ContentLength = int64(len(body))
	return &out
}

func RecorderMiddleware(rec *Recorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			body, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(body))
			recReq := req.Clone(req.Context())
			recReq.Body = io.NopCloser(bytes.NewReader(body))
			client := &http.Client{Transport: rec}
			proxyReq, err := http.NewRequestWithContext(req.Context(), recReq.Method, recReq.URL.String(), bytes.NewReader(body))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			proxyReq.Header = recReq.Header
			resp, err := client.Do(proxyReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			defer resp.Body.Close()
			for k, vv := range resp.Header {
				for _, v := range vv {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
		})
	}
}

func (r *Recorder) Flush(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return r.cassette.Flush()
}
