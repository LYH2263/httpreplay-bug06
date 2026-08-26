package httpreplay

import "errors"

var (
	ErrClosed     = errors.New("httpreplay: cassette closed")
	ErrNotFound   = errors.New("httpreplay: interaction not found")
	ErrMismatch   = errors.New("httpreplay: request mismatch")
	ErrInvalid    = errors.New("httpreplay: invalid argument")
	ErrSerialize  = errors.New("httpreplay: serialize failure")
	ErrNoMatch    = errors.New("httpreplay: no matching interaction")
	ErrReplayDone = errors.New("httpreplay: replay exhausted")
)
