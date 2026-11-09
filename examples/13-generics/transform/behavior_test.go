package main

import (
	"slices"
	"strconv"
	"testing"
)

func TestTransform(t *testing.T) {
	calls := 0
	got := transform([]int{2, 4}, func(n int) string { calls++; return strconv.Itoa(n) })
	if !slices.Equal(got, []string{"2", "4"}) || calls != 2 {
		t.Fatal(got, calls)
	}
	if transform[int, string](nil, strconv.Itoa) != nil {
		t.Fatal("nilness")
	}
	if got := transform([]int{}, strconv.Itoa); got == nil || len(got) != 0 {
		t.Fatal("empty nonnil")
	}
}
