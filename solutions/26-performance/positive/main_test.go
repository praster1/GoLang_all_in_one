package main

import (
	"slices"
	"testing"
)

func TestPositive(t *testing.T) {
	for _, tc := range []struct{ input, want []int }{
		{nil, nil}, {[]int{-1, 0}, nil}, {[]int{2, 2}, []int{2, 2}},
		{[]int{-2, 3, 0, 5, -1}, []int{3, 5}},
	} {
		before := slices.Clone(tc.input)
		for _, fn := range []func([]int) []int{positivesGrow, positivesReserve} {
			got := fn(tc.input)
			if !slices.Equal(got, tc.want) || !slices.Equal(tc.input, before) {
				t.Fatal("contract broken")
			}
			if len(got) > 0 {
				got[0] = 99
				if !slices.Equal(tc.input, before) {
					t.Fatal("shared result")
				}
			}
		}
	}
}

func BenchmarkPositive(b *testing.B) {
	for _, tc := range []struct {
		name  string
		every int
	}{{"All", 1}, {"Half", 2}, {"Sparse", 100}} {
		values := make([]int, 1000)
		for i := range values {
			if i%tc.every == 0 {
				values[i] = i + 1
			}
		}
		b.Run(tc.name, func(b *testing.B) {
			b.Run("Grow", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					positivesGrow(values)
				}
			})
			b.Run("Reserve", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					positivesReserve(values)
				}
			})
		})
	}
}

func Example_main() {
	main()
	// Output:
	// [3 5]
	// [3 5]
}
