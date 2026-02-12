package cond

import (
	"sync"
	"testing"
	"time"
)

func TestRestaurant_ConcurrentAccess(t *testing.T) {
	tables := 5

	r := NewRestaurant(tables)
	var wg sync.WaitGroup

	allOccupants := 7
	for range allOccupants {
		wg.Add(1)
		go func() {
			defer wg.Done()

			r.OccupyTable()
		}()
	}

	time.Sleep(500 * time.Millisecond)

	freeTables := r.GetAvailableTables()
	if freeTables != 0 {
		t.Fatalf("excpect 0 free tables, got: %v", freeTables)
	}

	realesed := 3
	for range realesed {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.ReleaseTable()
		}()
	}

	wg.Wait()

	freeTables = r.GetAvailableTables()
	expect := tables - (allOccupants - realesed)
	if freeTables != expect {
		t.Fatalf("excpect %v free tables after realesed, got: %v", expect, freeTables)
	}
}
