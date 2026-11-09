package main

import (
	"slices"
	"testing"
)

func TestFrequencyAndOrdering(t *testing.T) {
	counts := frequency([]string{"b", "a", "b", ""})
	if len(counts) != 3 || counts["b"] != 2 || counts["a"] != 1 || counts[""] != 1 {
		t.Fatal(counts)
	}
	if got := sortedKeys(counts); !slices.Equal(got, []string{"", "a", "b"}) {
		t.Fatal(got)
	}
	empty := frequency(nil)
	if empty == nil || len(empty) != 0 {
		t.Fatal(empty)
	}
	if len(sortedKeys(nil)) != 0 {
		t.Fatal("nil map keys")
	}
}

func Example_main() {
	main()
	// Output:
	// go: 3
	// map: 2
	// slice: 1
}
