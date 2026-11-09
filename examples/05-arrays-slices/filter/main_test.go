package main

import (
	"slices"
	"testing"
)

func TestAtLeast(t *testing.T) {
	for _, tc := range []struct {
		values []int
		min    int
		want   []int
	}{
		{nil, 0, nil}, {[]int{1, 2}, 3, nil}, {[]int{3, 2, 3}, 3, []int{3, 3}},
		{[]int{-2, 0, 5}, 0, []int{0, 5}},
	} {
		if got := atLeast(tc.values, tc.min); !slices.Equal(got, tc.want) {
			t.Fatalf("%v", got)
		}
	}
	original := []int{2, 3}
	selected := atLeast(original, 2)
	selected[0] = 8
	if original[0] != 2 {
		t.Fatal("result must be independent")
	}
}

func Example_main() {
	main()
	// Output:
	// 전체: [15 40 0 25 60]
	// 30분 이상: [40 60]
	// 결과 수정 뒤 원본: [15 40 0 25 60]
}
