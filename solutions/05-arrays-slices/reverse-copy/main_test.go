package main

import (
	"slices"
	"testing"
)

func TestReversed(t *testing.T) {
	for _, tc := range []struct{ in, want []int }{
		{nil, nil}, {[]int{}, nil}, {[]int{7}, []int{7}},
		{[]int{1, 2, 3}, []int{3, 2, 1}}, {[]int{1, 2, 2, 1}, []int{1, 2, 2, 1}},
	} {
		before := slices.Clone(tc.in)
		got := reversed(tc.in)
		if !slices.Equal(got, tc.want) || !slices.Equal(tc.in, before) || (len(tc.want) == 0 && got != nil) {
			t.Fatalf("in %v got %v", tc.in, got)
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
	// 뒤집은 복사: [40 30 20 10]
	// 원본: [10 20 30 40]
}
