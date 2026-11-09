package main

import (
	"slices"
	"testing"
)

func TestPassing(t *testing.T) {
	for _, tc := range []struct{ in, want []int }{
		{nil, nil}, {[]int{}, nil}, {[]int{0, 59}, nil},
		{[]int{60}, []int{60}}, {[]int{100, 60, 100}, []int{100, 60, 100}},
	} {
		if got := passing(tc.in); !slices.Equal(got, tc.want) || (len(tc.want) == 0 && got != nil) {
			t.Fatalf("got %v want %v", got, tc.want)
		}
	}
	in := []int{60, 90}
	out := passing(in)
	out[0] = 0
	if in[0] != 60 {
		t.Fatal("input changed through output")
	}
}

func Example_main() {
	main()
	// Output:
	// 통과 점수: [60 90 100]
	// 원본: [59 60 90 30 100]
}
