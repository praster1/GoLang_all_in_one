package main

import (
	"slices"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func TestTopGroups(t *testing.T) {
	source := report.Summary{Groups: []report.Group{{Category: "z", Total: 10}, {Category: "b", Total: 20}, {Category: "a", Total: 20}}}
	before := slices.Clone(source.Groups)
	got, err := topGroups(source, 2)
	if err != nil || len(got) != 2 || got[0].Category != "a" || got[1].Category != "b" {
		t.Fatalf("got=%v err=%v", got, err)
	}
	got[0].Total = 999
	if !slices.Equal(source.Groups, before) {
		t.Fatal("mutated source")
	}
	if got, err := topGroups(source, 0); err != nil || len(got) != 0 {
		t.Fatal("zero")
	}
	if got, err := topGroups(source, 9); err != nil || len(got) != 3 {
		t.Fatal("clamp")
	}
	if _, err := topGroups(source, -1); err == nil {
		t.Fatal("negative")
	}
	if got, err := topGroups(report.Summary{}, 2); err != nil || len(got) != 0 {
		t.Fatal("empty")
	}
}

func Example_main() {
	main()
	// Output:
	// 1. 도서: 20
	// 2. 식품: 20
}
