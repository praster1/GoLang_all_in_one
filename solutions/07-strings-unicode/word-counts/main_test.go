package main

import "testing"

func TestWordCounts(t *testing.T) {
	empty, ok := wordCounts(" \t\n")
	if !ok || empty == nil || len(empty) != 0 {
		t.Fatal(empty, ok)
	}
	bad, ok := wordCounts("good \xff")
	if ok || bad != nil {
		t.Fatal(bad, ok)
	}
	got, ok := wordCounts("GO go, 한글\u3000한글")
	if !ok || len(got) != 3 || got["go"] != 1 || got["go,"] != 1 || got["한글"] != 2 {
		t.Fatal(got, ok)
	}
}

func Example_main() {
	main()
	// Output:
	// "go": 2
	// "한글": 2
}
