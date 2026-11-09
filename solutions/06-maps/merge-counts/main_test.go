package main

import "testing"

func TestMergeCounts(t *testing.T) {
	if got := mergeCounts(nil, nil); got == nil || len(got) != 0 {
		t.Fatal(got)
	}
	a := map[string]int{"x": 2, "zero": 0}
	b := map[string]int{"x": 3, "y": 4}
	got := mergeCounts(a, b)
	if len(got) != 3 || got["x"] != 5 || got["y"] != 4 {
		t.Fatal(got)
	}
	if value, ok := got["zero"]; !ok || value != 0 {
		t.Fatal("zero entry lost")
	}
	got["x"] = 9
	delete(got, "zero")
	if a["x"] != 2 || b["x"] != 3 || len(a) != 2 {
		t.Fatal("input changed")
	}
	only := mergeCounts(nil, b)
	only["y"] = 8
	if b["y"] != 4 {
		t.Fatal("nil branch aliases input")
	}
}

func Example_main() {
	main()
	// Output:
	// go: 5
	// map: 1
	// slice: 2
	// 원본 go: 2 3
}
