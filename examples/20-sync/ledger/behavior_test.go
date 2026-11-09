package main

import (
	"reflect"
	"sync"
	"testing"
)

func TestTransfersPreserveInvariant(t *testing.T) {
	ledger := newLedger()
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 10 {
				if err := ledger.Transfer("A", "B", 1); err != nil {
					t.Error(err)
				}
				snapshot := ledger.Snapshot()
				if snapshot["A"]+snapshot["B"] != 200 || snapshot["A"] < 0 || snapshot["B"] < 0 {
					t.Error(snapshot)
				}
			}
		})
	}
	group.Wait()
	if got := ledger.Snapshot(); got["A"] != 20 || got["B"] != 180 {
		t.Fatal(got)
	}
}
func TestFailedTransferDoesNotChangeBalances(t *testing.T) {
	ledger := newLedger()
	before := ledger.Snapshot()
	for _, tt := range []struct {
		from, to string
		amount   int
	}{{"missing", "B", 1}, {"A", "missing", 1}, {"A", "A", 1}, {"A", "B", 0}, {"A", "B", 101}} {
		if err := ledger.Transfer(tt.from, tt.to, tt.amount); err == nil {
			t.Fatal(tt)
		}
		if !reflect.DeepEqual(ledger.Snapshot(), before) {
			t.Fatal("partial update")
		}
	}
	if err := ledger.Transfer("A", "B", 100); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Transfer("A", "B", 1); err == nil {
		t.Fatal("overdraft")
	}
	snapshot := ledger.Snapshot()
	snapshot["A"] = 999
	if ledger.Snapshot()["A"] != 0 {
		t.Fatal("leaked map")
	}
}
