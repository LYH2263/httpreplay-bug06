package httpreplay

import (
	"sync"
	"time"

	"github.com/LYH2263/go-httpreplay/internal/cassette"
	"github.com/LYH2263/go-httpreplay/internal/clone"
	"github.com/LYH2263/go-httpreplay/internal/serialize"
)

type Cassette struct {
	mu           sync.RWMutex
	name         string
	closed       bool
	interactions []Interaction
	store        *cassette.Store
	stats        Stats
}

func OpenCassette(opts Options) (*Cassette, error) {
	opts = opts.withDefaults()
	c := &Cassette{name: opts.Name}
	if opts.StorePath != "" {
		st, err := cassette.Open(opts.StorePath)
		if err != nil {
			return nil, err
		}
		c.store = st
		items, err := st.LoadAll()
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			c.interactions = append(c.interactions, fromStored(it))
		}
	}
	return c, nil
}

func (c *Cassette) Name() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.name
}

func (c *Cassette) Append(it Interaction) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	if it.ID == "" {
		it.ID = cassette.NewID(c.name, len(c.interactions))
	}
	it.RequestBody = clone.Bytes(it.RequestBody)
	it.ResponseBody = clone.Bytes(it.ResponseBody)
	c.interactions = append(c.interactions, it)
	c.stats.Recorded++
	if c.store != nil {
		if err := c.store.Append(toStored(it)); err != nil {
			c.interactions = c.interactions[:len(c.interactions)-1]
			c.stats.Recorded--
			return err
		}
	}
	return nil
}

func (c *Cassette) Interactions() []Interaction {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Interaction, len(c.interactions))
	for i, it := range c.interactions {
		out[i] = it
		out[i].RequestBody = clone.Bytes(it.RequestBody)
		out[i].ResponseBody = clone.Bytes(it.ResponseBody)
	}
	return out
}

func (c *Cassette) Find(id string) (Interaction, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, it := range c.interactions {
		if it.ID == id {
			return it, true
		}
	}
	return Interaction{}, false
}

func (c *Cassette) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := c.stats
	s.Pending = len(c.interactions)
	return s
}

func (c *Cassette) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make([]Interaction, len(c.interactions))
	for i, it := range c.interactions {
		items[i] = it
		items[i].RequestBody = clone.Bytes(it.RequestBody)
		items[i].ResponseBody = clone.Bytes(it.ResponseBody)
	}
	return Snapshot{
		Name:         c.name,
		Interactions: items,
		Stats:        c.stats,
	}
}

func (c *Cassette) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.store != nil {
		return c.store.Flush()
	}
	return nil
}

func (c *Cassette) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	if c.store != nil {
		return c.store.Close()
	}
	return nil
}

func (c *Cassette) CloseFlushCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0
	}
	n := 0
	if c.store != nil {
		n = c.store.PendingRecords()
		_ = c.store.Flush()
	}
	c.closed = true
	if c.store != nil {
		_ = c.store.Close()
	}
	return n
}

func (c *Cassette) bumpReplayed() {
	c.stats.Replayed++
}

func (c *Cassette) bumpMiss() {
	c.mu.Lock()
	c.stats.Misses++
	c.mu.Unlock()
}

func (c *Cassette) bumpUnmatched() {
	c.mu.Lock()
	c.stats.Unmatched++
	c.mu.Unlock()
}

func toStored(it Interaction) serialize.Record {
	return serialize.Record{
		ID:              it.ID,
		Method:          it.Method,
		URL:             it.URL,
		RequestHeaders:  it.RequestHeaders,
		RequestBody:     clone.Bytes(it.RequestBody),
		Status:          it.Status,
		ResponseHeaders: it.ResponseHeaders,
		ResponseBody:    clone.Bytes(it.ResponseBody),
		RecordedAt:      it.RecordedAt.UnixNano(),
	}
}

func fromStored(r serialize.Record) Interaction {
	return Interaction{
		ID:              r.ID,
		Method:          r.Method,
		URL:             r.URL,
		RequestHeaders:  r.RequestHeaders,
		RequestBody:     clone.Bytes(r.RequestBody),
		Status:          r.Status,
		ResponseHeaders: r.ResponseHeaders,
		ResponseBody:    clone.Bytes(r.ResponseBody),
		RecordedAt:      timeFromNano(r.RecordedAt),
	}
}

func timeFromNano(n int64) time.Time {
	if n == 0 {
		return time.Time{}
	}
	return time.Unix(0, n)
}
