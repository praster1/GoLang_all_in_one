package main

import "testing"

func TestCountWordsContract(t *testing.T) {
	if got := countWords(" \t\n"); got == nil || len(got) != 0 {
		t.Fatal(got)
	}
	got := countWords("GO go, 한글\u3000한글")
	if len(got) != 3 || got["go"] != 1 || got["go,"] != 1 || got["한글"] != 2 {
		t.Fatal(got)
	}
}

func Example_main() {
	main()
	// Output:
	// "go": 2
	// "한글": 2
}
