package cassette

import (
	"sync"

	"github.com/LYH2263/go-httpreplay/internal/serialize"
)

type Memory struct {
	mu    sync.RWMutex
	items []serialize.Record
}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Append(rec serialize.Record) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, rec)
}

func (m *Memory) All() []serialize.Record {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]serialize.Record, len(m.items))
	copy(out, m.items)
	return out
}

func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}
