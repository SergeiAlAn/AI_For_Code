// Package activity keeps an in-memory list of recently active users.
package activity

import "sync"

const DefaultLimit = 5

// Tracker stores the last `limit` activity marks. It is safe for concurrent use.
type Tracker struct {
	mu    sync.Mutex
	ids   []int64
	limit int
}

func NewTracker(limit int) *Tracker {
	if limit <= 0 {
		limit = DefaultLimit
	}
	return &Tracker{ids: make([]int64, 0, limit), limit: limit}
}

// Mark records the user as active, evicting the oldest mark when full,
// and returns a snapshot of the list.
func (t *Tracker) Mark(id int64) []int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.ids) == t.limit {
		copy(t.ids, t.ids[1:])
		t.ids = t.ids[:len(t.ids)-1]
	}
	t.ids = append(t.ids, id)
	return t.snapshot()
}

func (t *Tracker) List() []int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.snapshot()
}

func (t *Tracker) snapshot() []int64 {
	return append(make([]int64, 0, len(t.ids)), t.ids...)
}
