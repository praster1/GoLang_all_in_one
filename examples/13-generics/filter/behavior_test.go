package main

import (
	"slices"
	"testing"
)

func TestFilter(t *testing.T) {
	source := Scores{1, 2, 3, 4}
	calls := 0
	got := filter(source, func(v int) bool { calls++; return v%2 == 0 })
	if !slices.Equal(got, Scores{2, 4}) || calls != 4 {
		t.Fatal(got, calls)
	}
	got[0] = 9
	if source[1] != 2 {
		t.Fatal("shares outer backing array")
	}
	if filter(Scores(nil), func(int) bool { return true }) != nil {
		t.Fatal("nilness")
	}
	if got := filter(Scores{}, func(int) bool { return true }); got == nil || len(got) != 0 {
		t.Fatal("empty nonnil contract")
	}
	if got := filter(source, func(int) bool { return false }); len(got) != 0 || got == nil {
		t.Fatal("no match")
	}
}
