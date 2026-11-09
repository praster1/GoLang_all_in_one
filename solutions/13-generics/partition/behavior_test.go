package main

import (
	"slices"
	"testing"
)

type Numbers []int

func TestPartition(t *testing.T) {
	input := Numbers{1, 2, 3, 4}
	calls := 0
	yes, no := partition(input, func(v int) bool { calls++; return v%2 == 0 })
	if calls != 4 || !slices.Equal(yes, Numbers{2, 4}) || !slices.Equal(no, Numbers{1, 3}) {
		t.Fatal(yes, no, calls)
	}
	yes[0] = 99
	no[0] = 88
	if !slices.Equal(input, Numbers{1, 2, 3, 4}) {
		t.Fatal(input)
	}
	a, b := partition(Numbers(nil), func(int) bool { return true })
	if a != nil || b != nil {
		t.Fatal("nilness")
	}
	a, b = partition(Numbers{}, func(int) bool { return true })
	if a == nil || b == nil || len(a)+len(b) != 0 {
		t.Fatal("empty nonnil")
	}
}
