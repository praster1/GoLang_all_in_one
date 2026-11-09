package main

import (
	"slices"
	"testing"
)

func TestUniqueNames(t *testing.T) {
	if uniqueNames(nil) != nil {
		t.Fatal("nil result")
	}
	in := []string{"b", "a", "b", "", "a", ""}
	got := uniqueNames(in)
	if !slices.Equal(got, []string{"b", "a", ""}) {
		t.Fatal(got)
	}
	got[0] = "changed"
	if in[0] != "b" {
		t.Fatal("shared result")
	}
}

func Example_main() {
	main()
	// Output:
	// 처음 등장 순서: [Mina Jin Sol]
	// 원본: [Mina Jin Mina Sol Jin]
}
