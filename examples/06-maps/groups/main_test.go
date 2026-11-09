package main

import (
	"slices"
	"testing"
)

func TestGroupParity(t *testing.T) {
	input := []int{-2, -1, 0, 3}
	got := groupParity(input)
	if !slices.Equal(got["even"], []int{-2, 0}) || !slices.Equal(got["odd"], []int{-1, 3}) {
		t.Fatal(got)
	}
	got["even"][0] = 9
	if input[0] != -2 {
		t.Fatal("shared input")
	}
	if empty := groupParity(nil); empty == nil || len(empty) != 0 {
		t.Fatal(empty)
	}
}

func Example_main() {
	main()
	// Output:
	// 짝수: [2 4 0]
	// 홀수: [1 5]
	// 없는 그룹 길이: 0
}
