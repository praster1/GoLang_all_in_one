package main

import (
	"errors"
	"testing"
)

func FuzzClamp(f *testing.F) {
	f.Add(120, 0, 100)
	f.Add(-1, 0, 100)
	f.Add(5, 5, 5)
	f.Add(1, 10, 0)
	f.Fuzz(func(t *testing.T, value, low, high int) {
		got, err := clamp(value, low, high)
		if low > high {
			if !errors.Is(err, errBounds) {
				t.Fatal("bad bounds accepted")
			}
			return
		}
		if err != nil || got < low || got > high {
			t.Fatal("outside range")
		}
		if value >= low && value <= high && got != value {
			t.Fatal("changed in-range input")
		}
		again, err := clamp(got, low, high)
		if err != nil || again != got {
			t.Fatal("not idempotent")
		}
	})
}

func TestClampEdges(t *testing.T) {
	max := int(^uint(0) >> 1)
	min := -max - 1
	got, err := clamp(max, min, max)
	if err != nil || got != max {
		t.Fatal("integer edge")
	}
}

func Example_main() {
	main()
	// Output:
	// 보정: 100
}
