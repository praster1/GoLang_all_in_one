package main

import (
	"slices"
	"testing"
)

func TestCloneIntsIndependence(t *testing.T) {
	original := []int{1, 2, 3}
	got := cloneInts(original)
	if !slices.Equal(got, original) {
		t.Fatal(got)
	}
	got[0] = 7
	if original[0] != 1 {
		t.Fatal("copy shares elements")
	}
	original[1] = 8
	if got[1] != 2 {
		t.Fatal("reverse alias")
	}
}
func TestCloneIntsEmpty(t *testing.T) {
	if cloneInts(nil) != nil {
		t.Fatal("nil must be preserved")
	}
	if got := cloneInts([]int{}); got == nil || len(got) != 0 {
		t.Fatalf("%#v", got)
	}
}

func Example_main() {
	main()
	// Output:
	// 원본: [1 2 3]
	// 독립 복사: [99 2 3]
	// 짧은 목적지: [1 2], 복사 수: 2
	// 겹치는 복사: [1 1 2 3]
}
