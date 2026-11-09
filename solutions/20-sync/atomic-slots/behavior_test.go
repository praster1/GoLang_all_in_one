package main

import (
	"sync"
	"testing"
)

func TestSlotsBoundAndIdempotentRelease(t *testing.T) {
	slots, err := newSlots(2)
	if err != nil {
		t.Fatal(err)
	}
	first, ok := slots.TryAcquire()
	if !ok {
		t.Fatal("first")
	}
	second, ok := slots.TryAcquire()
	if !ok {
		t.Fatal("second")
	}
	if release, ok := slots.TryAcquire(); ok || release != nil {
		t.Fatal("limit")
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(first)
	}
	group.Wait()
	if slots.active.Load() != 1 {
		t.Fatal(slots.active.Load())
	}
	second()
	if slots.active.Load() != 0 {
		t.Fatal(slots.active.Load())
	}
	for range 16 {
		group.Go(func() {
			for range 100 {
				if release, ok := slots.TryAcquire(); ok {
					if n := slots.active.Load(); n < 1 || n > 2 {
						t.Error(n)
					}
					release()
				}
			}
		})
	}
	group.Wait()
	if slots.active.Load() != 0 {
		t.Fatal(slots.active.Load())
	}
	for _, limit := range []int64{0, 33} {
		if value, err := newSlots(limit); value != nil || err == nil {
			t.Fatal(value, err)
		}
	}
}
