package main

import (
	"slices"
	"testing"
)

func TestUnique(t *testing.T) {
	input := []int{2, 1, 2, 3}
	got := unique(input)
	if !slices.Equal(got, []int{2, 1, 3}) {
		t.Fatal(got)
	}
	got[0] = 99
	if input[0] != 2 {
		t.Fatal("mutates source")
	}
	if unique([]int(nil)) != nil {
		t.Fatal("nilness")
	}
	if got := unique([]int{}); got == nil || len(got) != 0 {
		t.Fatal("empty nonnil")
	}
	var names Names = unique(Names{"a", "a"})
	if len(names) != 1 {
		t.Fatal(names)
	}
}
