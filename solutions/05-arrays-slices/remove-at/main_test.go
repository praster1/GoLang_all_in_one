package main

import (
	"slices"
	"testing"
)

func TestRemoveAt(t *testing.T) {
	for _, tc := range []struct {
		in    []int
		index int
		want  []int
		ok    bool
	}{
		{nil, 0, nil, false}, {[]int{1}, -1, nil, false}, {[]int{1}, 1, nil, false},
		{[]int{1}, 0, []int{}, true}, {[]int{1, 2, 3}, 0, []int{2, 3}, true},
		{[]int{1, 2, 3}, 1, []int{1, 3}, true}, {[]int{1, 2, 3}, 2, []int{1, 2}, true},
	} {
		before := slices.Clone(tc.in)
		got, ok := removeAt(tc.in, tc.index)
		if ok != tc.ok || !slices.Equal(got, tc.want) || !slices.Equal(tc.in, before) {
			t.Fatalf("got %v,%t", got, ok)
		}
		if !ok && got != nil {
			t.Fatal("invalid result must be nil")
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
	// 삭제 성공: true
	// 결과: [10 30 40]
	// 원본: [10 20 30 40]
}
