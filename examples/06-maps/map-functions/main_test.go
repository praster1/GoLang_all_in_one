package main

import "testing"

func TestMapMutationAndClone(t *testing.T) {
	if increment(nil, "x") {
		t.Fatal("nil must fail")
	}
	input := map[string]int{"x": 0}
	if !increment(input, "x") || input["x"] != 1 {
		t.Fatal(input)
	}
	if !increment(input, "new") || input["new"] != 1 {
		t.Fatal(input)
	}
	copied := cloneCounts(input)
	copied["x"] = 9
	delete(copied, "new")
	if input["x"] != 1 || input["new"] != 1 {
		t.Fatal("clone changed source")
	}
	if cloneCounts(nil) != nil {
		t.Fatal("nil clone")
	}
	if c := cloneCounts(map[string]int{}); c == nil || len(c) != 0 {
		t.Fatal(c)
	}
}

func Example_main() {
	main()
	// Output:
	// 변경: 3, 성공: true
	// 원본: 3
	// 복사: 99
	// nil 변경: false
}
