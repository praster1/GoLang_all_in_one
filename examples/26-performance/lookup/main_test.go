package main

import "testing"

func TestLookup(t *testing.T) {
	values := []int{1, 1, 3, 9}
	set := makeSet(values)
	if len(set) != 3 {
		t.Fatal("duplicate key")
	}
	for _, target := range []int{0, 1, 3, 9, 10} {
		_, got := set[target]
		if got != contains(values, target) {
			t.Fatalf("target=%d", target)
		}
	}
	if contains(nil, 0) {
		t.Fatal("empty input")
	}
}

func BenchmarkLookup(b *testing.B) {
	values := make([]int, 1000)
	for i := range values {
		values[i] = i
	}
	set := makeSet(values)
	b.Run("LinearMissing", func(b *testing.B) {
		for b.Loop() {
			contains(values, -1)
		}
	})
	b.Run("MapMissingPrepared", func(b *testing.B) {
		for b.Loop() {
			containsSet(set, -1)
		}
	})
	b.Run("MapBuildAndLookup", func(b *testing.B) {
		for b.Loop() {
			current := makeSet(values)
			containsSet(current, -1)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// true true
	// false false
}
