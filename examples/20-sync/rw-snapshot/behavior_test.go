package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestSnapshotsAreIndependent(t *testing.T) {
	var store Store
	var group sync.WaitGroup
	for i := range 8 {
		group.Go(func() {
			key := fmt.Sprint(i)
			for range 50 {
				store.Put(key, "ok")
				snapshot := store.Snapshot()
				snapshot["external"] = "change"
			}
		})
	}
	group.Wait()
	snapshot := store.Snapshot()
	if len(snapshot) != 8 {
		t.Fatal(snapshot)
	}
	if _, exists := snapshot["external"]; exists {
		t.Fatal("map escaped")
	}
}
