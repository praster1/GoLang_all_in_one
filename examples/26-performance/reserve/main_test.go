package main

import (
	"slices"
	"testing"
)

func TestSquares(t *testing.T) {
	for _, n := range []int{-1, 0, 1, 5, 100} {
		a, b := squaresGrow(n), squaresReserve(n)
		if !slices.Equal(a, b) {
			t.Fatalf("n=%d: %v != %v", n, a, b)
		}
		for i, value := range b {
			if value != i*i {
				t.Fatal("incorrect square")
			}
		}
	}
}

func BenchmarkSquares(b *testing.B) {
	for _, n := range []int{10, 1000} {
		b.Run(fmtSize(n), func(b *testing.B) {
			b.Run("Grow", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					squaresGrow(n)
				}
			})
			b.Run("Reserve", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					squaresReserve(n)
				}
			})
		})
	}
}

func fmtSize(n int) string {
	if n == 10 {
		return "N10"
	}
	return "N1000"
}

func Example_main() {
	main()
	// Output:
	// [0 1 4 9 16]
	// [0 1 4 9 16]
}
