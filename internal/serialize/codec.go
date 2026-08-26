package serialize

import (
	"encoding/json"
	"io"
)

func WriteRecords(w io.Writer, recs []Record) error {
	enc := json.NewEncoder(w)
	for _, r := range recs {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

func ReadRecords(r io.Reader) ([]Record, error) {
	dec := json.NewDecoder(r)
	var out []Record
	for {
		var rec Record
		if err := dec.Decode(&rec); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}
