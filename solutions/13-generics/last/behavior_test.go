package main

import "testing"

func TestLast(t *testing.T) {
	if got, ok := last([]int{0}); !ok || got != 0 {
		t.Fatal(got, ok)
	}
	if got, ok := last[int](nil); ok || got != 0 {
		t.Fatal(got, ok)
	}
	if got, ok := last([]string{"a", "b"}); !ok || got != "b" {
		t.Fatal(got, ok)
	}
}
