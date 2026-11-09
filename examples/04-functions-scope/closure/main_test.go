package main

import "testing"

func TestNewCounterIndependentCounters(t *testing.T) {
	first := newCounter()
	second := newCounter()

	for want := 1; want <= 3; want++ {
		if got := first(); got != want {
			t.Fatalf("first() = %d, want %d", got, want)
		}
		if got := second(); got != want {
			t.Fatalf("second() = %d, want %d", got, want)
		}
	}
}

func TestNewCounterAliasSharesState(t *testing.T) {
	first := newCounter()
	alias := first

	if got := first(); got != 1 {
		t.Fatalf("first() = %d, want 1", got)
	}
	if got := alias(); got != 2 {
		t.Fatalf("alias() = %d, want 2", got)
	}
	if got := first(); got != 3 {
		t.Fatalf("first() after alias() = %d, want 3", got)
	}
}

func Example_main() {
	main()
	// Output:
	// 첫 번째: 1
	// 첫 번째: 2
	// 두 번째: 1
	// 첫 번째: 3
}
