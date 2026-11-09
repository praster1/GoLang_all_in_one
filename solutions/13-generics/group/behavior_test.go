package main

import (
	"slices"
	"testing"
)

func TestGroup(t *testing.T) {
	input := []int{1, 2, 3, 4}
	calls := 0
	got := groupBy(input, func(v int) bool { calls++; return v%2 == 0 })
	if !slices.Equal(got[true], []int{2, 4}) || !slices.Equal(got[false], []int{1, 3}) || calls != 4 {
		t.Fatal(got, calls)
	}
	got[false][0] = 99
	if input[0] != 1 {
		t.Fatal("shares source")
	}
	empty := groupBy([]int(nil), func(int) string { return "x" })
	if empty == nil || len(empty) != 0 {
		t.Fatal("empty map contract")
	}
}
