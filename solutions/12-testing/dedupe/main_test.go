package main

import (
	"slices"
	"testing"
)

func TestUnique(t *testing.T) {
	for _, tc := range []struct {
		name        string
		input, want []string
	}{
		{"nil", nil, []string{}}, {"empty", []string{}, []string{}},
		{"order", []string{"b", "a", "b", "c"}, []string{"b", "a", "c"}},
		{"case", []string{"Go", "go", "Go"}, []string{"Go", "go"}},
		{"empty_string", []string{"", "x", ""}, []string{"", "x"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := slices.Clone(tc.input)
			got := unique(tc.input)
			if got == nil || !slices.Equal(got, tc.want) {
				t.Fatalf("got=%v; want %v", got, tc.want)
			}
			if len(got) > 0 {
				got[0] = "changed"
			}
			if !slices.Equal(tc.input, before) {
				t.Error("input changed")
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// [Go 한글 go]
}
