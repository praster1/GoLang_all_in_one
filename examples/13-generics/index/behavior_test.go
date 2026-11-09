package main

import "testing"

func TestIndex(t *testing.T) {
	for _, tc := range []struct {
		values       []int
		target, want int
	}{{nil, 1, -1}, {[]int{1, 2, 1}, 1, 0}, {[]int{1, 2}, 2, 1}, {[]int{1}, 2, -1}} {
		if got := index(tc.values, tc.target); got != tc.want {
			t.Fatal(tc, got)
		}
	}
	if index([]Code{2, 3}, Code(3)) != 1 {
		t.Fatal("defined type")
	}
}
