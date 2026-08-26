package normalize

import (
	"bytes"
	"encoding/json"
)

func NormalizeBody(body []byte, cfg Config) []byte {
	if len(body) == 0 {
		return nil
	}
	if !json.Valid(body) {
		return append([]byte(nil), body...)
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return append([]byte(nil), body...)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return append([]byte(nil), body...)
	}
	return bytes.TrimSpace(b)
}
