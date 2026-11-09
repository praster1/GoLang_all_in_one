package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestReserveNeverOvershoots(t *testing.T) {
	quota := Quota{limit: 101}
	var count atomic.Int64
	var group sync.WaitGroup
	for range 16 {
		group.Go(func() {
			for range 20 {
				if quota.Reserve(2) {
					count.Add(1)
				}
			}
		})
	}
	group.Wait()
	if quota.used.Load() != 100 || count.Load() != 50 {
		t.Fatal(quota.used.Load(), count.Load())
	}
	for _, amount := range []int64{-1, 0, 2, 102} {
		if quota.Reserve(amount) {
			t.Fatal(amount)
		}
	}
	if !quota.Reserve(1) || quota.used.Load() != 101 {
		t.Fatal(quota.used.Load())
	}
}
