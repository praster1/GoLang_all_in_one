package main

import "testing"

func TestReplaceFirstCopiesArray(t *testing.T) {
	original := [3]int{1, 2, 3}
	got := replaceFirst(original)
	if original != [3]int{1, 2, 3} || got != [3]int{99, 2, 3} {
		t.Fatalf("original=%v, got=%v", original, got)
	}
	if got := replaceFirst([3]int{}); got != [3]int{99, 0, 0} {
		t.Fatal(got)
	}
}

func Example_main() {
	main()
	// Output:
	// 초기: [20 30 0]
	// 원본: [20 30 0]
	// 복사본: [99 30 0]
	// 길이: 3
}
