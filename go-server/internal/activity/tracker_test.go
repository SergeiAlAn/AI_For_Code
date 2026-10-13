package activity

import (
	"slices"
	"sync"
	"testing"
)

func TestTrackerEvictsOldest(t *testing.T) {
	tr := NewTracker(3)
	if got := tr.List(); got == nil || len(got) != 0 {
		t.Fatalf("empty tracker: got %#v", got)
	}
	for id := int64(1); id <= 5; id++ {
		tr.Mark(id)
	}
	if got := tr.List(); !slices.Equal(got, []int64{3, 4, 5}) {
		t.Fatalf("got %v", got)
	}
	if got := tr.Mark(1); !slices.Equal(got, []int64{4, 5, 1}) {
		t.Fatalf("got %v", got)
	}
}

func TestTrackerSnapshotIsCopy(t *testing.T) {
	tr := NewTracker(2)
	snap := tr.Mark(1)
	snap[0] = 99
	if got := tr.List(); got[0] != 1 {
		t.Fatalf("tracker state changed through snapshot: %v", got)
	}
}

func TestTrackerConcurrent(t *testing.T) {
	tr := NewTracker(DefaultLimit)
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tr.Mark(int64(i))
		}()
	}
	wg.Wait()
	if got := tr.List(); len(got) != DefaultLimit {
		t.Fatalf("got %d marks, want %d", len(got), DefaultLimit)
	}
}
