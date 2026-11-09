package main

import (
	"errors"
	"testing"
)

func TestClamp(t *testing.T) {
	tests := []struct {
		name                  string
		value, min, max, want int
	}{
		{"below", -1, 0, 100, 0}, {"minimum", 0, 0, 100, 0}, {"inside", 50, 0, 100, 50},
		{"maximum", 100, 0, 100, 100}, {"above", 101, 0, 100, 100}, {"one_point", 20, 7, 7, 7},
		{"negative_range", -5, -10, -2, -5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := clamp(tc.value, tc.min, tc.max)
			if err != nil || got != tc.want {
				t.Fatalf("got=%d err=%v; want %d", got, err, tc.want)
			}
		})
	}
}

func TestClampInvalid(t *testing.T) {
	if got, err := clamp(5, 10, 0); got != 0 || !errors.Is(err, errBounds) {
		t.Fatalf("got=%d err=%v", got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 보정: 100
}
