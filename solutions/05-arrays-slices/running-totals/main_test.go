package main

import (
	"slices"
	"testing"
)

func TestRunningTotals(t *testing.T) {
	for _, tc := range []struct{ in, want []int }{
		{nil, nil}, {[]int{}, nil}, {[]int{5}, []int{5}},
		{[]int{20, 0, 30, 15}, []int{20, 20, 50, 65}},
		{[]int{5, -5, 2}, []int{5, 0, 2}},
	} {
		before := slices.Clone(tc.in)
		got := runningTotals(tc.in)
		if !slices.Equal(got, tc.want) || !slices.Equal(tc.in, before) || (len(tc.want) == 0 && got != nil) {
			t.Fatalf("%v", got)
		}
		if len(got) > 0 {
			got[0] = 99
			if !slices.Equal(tc.in, before) {
				t.Fatal("shared result")
			}
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 일별: [20 0 30 15]
	// 누적: [20 20 50 65]
}
