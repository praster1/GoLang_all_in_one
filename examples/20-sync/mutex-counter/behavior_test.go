package main

import (
	"sync"
	"testing"
)

func TestCounterConcurrentReadersAndWriters(t *testing.T) {
	var counter Counter
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 100 {
				counter.Add(1)
				if counter.Value() < 0 {
					t.Error("negative value")
				}
			}
		})
	}
	group.Wait()
	if got := counter.Value(); got != 800 {
		t.Fatal(got)
	}
}
