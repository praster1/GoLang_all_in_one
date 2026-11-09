package main

import "testing"

func TestTally(t *testing.T) {
	if got := tally(nil); got == nil || len(got) != 0 {
		t.Fatal(got)
	}
	got := tally([]string{"a", "a", "A", "", ""})
	if len(got) != 3 || got["a"] != 2 || got["A"] != 1 || got[""] != 2 {
		t.Fatal(got)
	}
	empty := tally([]string{})
	if empty == nil || len(empty) != 0 {
		t.Fatal(empty)
	}
}

func Example_main() {
	main()
	// Output:
	// book: 2
	// note: 1
	// pen: 3
}
