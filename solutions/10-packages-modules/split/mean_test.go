package main

import "testing"

func TestMean(t *testing.T) {
	for _, tc := range []struct {
		scores []int
		want   float64
		ok     bool
	}{{nil, 0, false}, {[]int{0}, 0, true}, {[]int{100}, 100, true}, {[]int{10, 20, 30}, 20, true}, {[]int{-1}, 0, false}, {[]int{101}, 0, false}, {make([]int, 101), 0, false}} {
		got, ok := mean(tc.scores)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("%+v: %v %t", tc, got, ok)
		}
	}
}
