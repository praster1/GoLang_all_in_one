package main

import "testing"

func TestDivide(t *testing.T) {
	for _, tc := range []struct {
		a, b, want int
		bad        bool
	}{{12, 3, 4, false}, {0, 1, 0, false}, {12, 0, 0, true}, {12, -1, 0, true}, {-5, 2, -2, false}} {
		got, err := divide(tc.a, tc.b)
		if got != tc.want || (err != nil) != tc.bad {
			t.Fatalf("%+v: %d %v", tc, got, err)
		}
	}
}
