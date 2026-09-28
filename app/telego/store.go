package telego

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const defaultCapacity = 1000

type Entry struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Name         string         `json:"name"`
	TraceID      string         `json:"trace_id,omitempty"`
	SpanID       string         `json:"span_id,omitempty"`
	ParentSpanID string         `json:"parent_span_id,omitempty"`
	Method       string         `json:"method,omitempty"`
	URI          string         `json:"uri,omitempty"`
	Status       string         `json:"status"`
	StatusCode   int            `json:"status_code,omitempty"`
	Duration     int64          `json:"duration_us"`
	Timestamp    time.Time      `json:"timestamp"`
	Payload      map[string]any `json:"payload,omitempty"`
}

type Filters struct {
	Type   string
	Search string
	Limit  int
}

type Store struct {
	mu       sync.RWMutex
	entries  []Entry
	capacity int
	paused   bool
	sequence atomic.Uint64
}

var defaultStore = NewStore(defaultCapacity)
var enabled atomic.Bool

func init() {
	enabled.Store(true)
}

func DefaultStore() *Store {
	return defaultStore
}

func SetEnabled(value bool) {
	enabled.Store(value)
}

func Enabled() bool {
	return enabled.Load()
}

func NewStore(capacity int) *Store {
	if capacity <= 0 {
		capacity = defaultCapacity
	}

	return &Store{
		entries:  make([]Entry, 0, capacity),
		capacity: capacity,
	}
}

func (r *Store) Add(entry Entry) Entry {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.paused {
		return entry
	}

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%x-%x", time.Now().UnixNano(), r.sequence.Add(1))
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if entry.Status == "" {
		entry.Status = "ok"
	}

	r.entries = append(r.entries, Entry{})
	copy(r.entries[1:], r.entries[:len(r.entries)-1])
	r.entries[0] = entry
	if len(r.entries) > r.capacity {
		r.entries = r.entries[:r.capacity]
	}

	return entry
}

func (r *Store) List(filters Filters) []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	limit := filters.Limit
	if limit <= 0 || limit > r.capacity {
		limit = 100
	}
	search := strings.ToLower(strings.TrimSpace(filters.Search))
	entries := make([]Entry, 0, min(limit, len(r.entries)))

	for _, entry := range r.entries {
		if filters.Type != "" && entry.Type != filters.Type {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(entry.Name+" "+entry.Method+" "+entry.URI+" "+entry.TraceID), search) {
			continue
		}

		entries = append(entries, entry)
		if len(entries) == limit {
			break
		}
	}

	return entries
}

func (r *Store) Get(id string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, entry := range r.entries {
		if entry.ID == id {
			return entry, true
		}
	}

	return Entry{}, false
}

func (r *Store) Update(id string, update func(entry *Entry)) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index := range r.entries {
		if r.entries[index].ID == id {
			entry := r.entries[index]
			update(&entry)
			r.entries[index] = entry
			return true
		}
	}

	return false
}

func (r *Store) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries = r.entries[:0]
}

func (r *Store) SetPaused(paused bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.paused = paused
}

func (r *Store) Paused() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.paused
}

func (r *Store) Stats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := map[string]int{"all": len(r.entries)}
	for _, entry := range r.entries {
		stats[entry.Type]++
	}

	return stats
}
