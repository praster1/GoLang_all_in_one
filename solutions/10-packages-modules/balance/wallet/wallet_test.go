package wallet

import "testing"

func TestWallet(t *testing.T) {
	for _, initial := range []int{-1, 10001} {
		if _, ok := New(initial); ok {
			t.Fatalf("accepted %d", initial)
		}
	}
	for _, initial := range []int{0, 10000} {
		w, ok := New(initial)
		if !ok || w.Balance() != initial {
			t.Fatal(initial)
		}
	}
	w, _ := New(100)
	if w.Spend(-1) || w.Spend(101) || w.Balance() != 100 {
		t.Fatal("invalid spend")
	}
	if !w.Spend(100) || w.Balance() != 0 {
		t.Fatal("exact spend")
	}
	var missing *Wallet
	if missing.Spend(0) {
		t.Fatal("nil accepted")
	}
}
