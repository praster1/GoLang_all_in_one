package main

import (
	"math"
	"testing"
)

func TestSmallest(t *testing.T) {
	if n, ok := smallest([]int{5, 1, 3}); !ok || n != 1 {
		t.Fatal(n, ok)
	}
	if _, ok := smallest([]string(nil)); ok {
		t.Fatal("empty accepted")
	}
	for _, v := range [][]float64{{2, math.NaN(), 1}, {math.NaN(), 1}, {1, math.NaN()}} {
		got, ok := smallest(v)
		if !ok || !math.IsNaN(got) {
			t.Fatal(v, got)
		}
	}
}
