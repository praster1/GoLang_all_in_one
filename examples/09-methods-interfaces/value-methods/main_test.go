package main

import "testing"

func TestAdded(t *testing.T) {
	original := Score{Points: 70}
	got := original.Added(-5)
	if original.Points != 70 || got.Points != 65 || got.Label() != "65점" {
		t.Fatalf("original=%+v, changed=%+v", original, got)
	}
}

func Example_main() {
	main()
	// Output:
	// 원본: 70점
	// 새 값: 75점
}
