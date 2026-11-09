package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestReserveConcurrent(t *testing.T) {
	inventory, err := newInventory(1000)
	if err != nil {
		t.Fatal(err)
	}
	var count atomic.Int64
	var group sync.WaitGroup
	for range 20 {
		group.Go(func() {
			for range 100 {
				if inventory.Reserve(7) {
					count.Add(1)
				}
			}
		})
	}
	group.Wait()
	if count.Load() != 142 || inventory.Remaining() != 6 {
		t.Fatal(count.Load(), inventory.Remaining())
	}
	for _, quantity := range []int{-1, 0, 7, 101} {
		if inventory.Reserve(quantity) {
			t.Fatal(quantity)
		}
	}
	if inventory.Remaining() != 6 {
		t.Fatal("changed on failure")
	}
}
func TestInventoryBounds(t *testing.T) {
	for _, stock := range []int{-1, 10001} {
		if value, err := newInventory(stock); value != nil || err == nil {
			t.Fatal(value, err)
		}
	}
}
