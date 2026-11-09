package main

import "testing"

func TestFirst(t *testing.T) {
	if got, ok := first[int](nil); ok || got != 0 {
		t.Fatal(got, ok)
	}
	if got, ok := first([]int{0}); !ok || got != 0 {
		t.Fatal(got, ok)
	}
	got, ok := first([][]int{{1}})
	if !ok || len(got) != 1 || got[0] != 1 {
		t.Fatal(got, ok)
	}
}
