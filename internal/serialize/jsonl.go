package serialize

import (
	"encoding/json"
)

type Record struct {
	ID              string              `json:"id"`
	Method          string              `json:"method"`
	URL             string              `json:"url"`
	RequestHeaders  map[string][]string `json:"request_headers,omitempty"`
	RequestBody     []byte              `json:"request_body,omitempty"`
	Status          int                 `json:"status"`
	ResponseHeaders map[string][]string `json:"response_headers,omitempty"`
	ResponseBody    []byte              `json:"response_body,omitempty"`
	RecordedAt      int64               `json:"recorded_at"`
}

func EncodeLine(r Record) ([]byte, error) {
	return json.Marshal(r)
}

func DecodeLine(b []byte) (Record, error) {
	var r Record
	err := json.Unmarshal(b, &r)
	return r, err
}
