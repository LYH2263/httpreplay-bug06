package cassette

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/LYH2263/go-httpreplay/internal/serialize"
)

type Store struct {
	mu      sync.Mutex
	path    string
	f       *os.File
	w       *bufio.Writer
	queue   []serialize.Record
	pending int
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	return &Store{path: path, f: f, w: bufio.NewWriter(f)}, nil
}

func (s *Store) Append(rec serialize.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec.Status < 0 {
		return fmt.Errorf("store reject status %d", rec.Status)
	}
	if rec.RequestBody != nil {
		rec.RequestBody = append([]byte(nil), rec.RequestBody...)
	}
	if rec.ResponseBody != nil {
		rec.ResponseBody = append([]byte(nil), rec.ResponseBody...)
	}
	s.queue = append(s.queue, rec)
	s.pending++
	return nil
}

func (s *Store) PendingRecords() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pending
}

func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range s.queue {
		b, err := serialize.EncodeLine(rec)
		if err != nil {
			return err
		}
		if _, err := s.w.Write(b); err != nil {
			return err
		}
		if err := s.w.WriteByte('\n'); err != nil {
			return err
		}
	}
	s.queue = nil
	if err := s.w.Flush(); err != nil {
		return err
	}
	s.pending = 0
	return s.f.Sync()
}

func (s *Store) LoadAll() ([]serialize.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f == nil {
		return nil, nil
	}
	if _, err := s.f.Seek(0, 0); err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(s.f)
	var out []serialize.Record
	for sc.Scan() {
		rec, err := serialize.DecodeLine(sc.Bytes())
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, sc.Err()
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range s.queue {
		b, _ := serialize.EncodeLine(rec)
		_, _ = s.w.Write(b)
		_ = s.w.WriteByte('\n')
	}
	s.queue = nil
	_ = s.w.Flush()
	if s.f != nil {
		return s.f.Close()
	}
	return nil
}

func (s *Store) Rotate(newPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range s.queue {
		b, err := serialize.EncodeLine(rec)
		if err != nil {
			return err
		}
		if _, err := s.w.Write(b); err != nil {
			return err
		}
		if err := s.w.WriteByte('\n'); err != nil {
			return err
		}
	}
	s.queue = nil
	if err := s.w.Flush(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(newPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	old := s.f
	if err := old.Close(); err != nil {
		_ = f.Close()
		return err
	}
	s.path = newPath
	s.f = f
	s.w = bufio.NewWriter(f)
	return nil
}

func (s *Store) Path() string { return s.path }
