package main

import "testing"

func TestSwap(t *testing.T) {
	left, right := -2, 7
	if !swap(&left, &right) || left != 7 || right != -2 {
		t.Fatalf("swap: %d %d", left, right)
	}
	if !swap(&left, &left) || left != 7 {
		t.Fatalf("same pointer: %d", left)
	}
	if swap(nil, &left) || swap(&left, nil) || swap(nil, nil) || left != 7 {
		t.Fatal("nil handling changed value")
	}
}

func Example_main() {
	main()
	// Output:
	// 교환: true
	// 값: 8 3
	// nil: false
}
