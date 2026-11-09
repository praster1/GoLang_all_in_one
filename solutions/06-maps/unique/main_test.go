package main

import (
	"slices"
	"testing"
)

func TestUnique(t *testing.T) {
	for _, tc := range []struct{ in, want []string }{
		{nil, nil}, {[]string{}, nil}, {[]string{"a"}, []string{"a"}},
		{[]string{"b", "a", "b", "A", "", ""}, []string{"b", "a", "A", ""}},
	} {
		before := slices.Clone(tc.in)
		got := unique(tc.in)
		if !slices.Equal(got, tc.want) || !slices.Equal(tc.in, before) {
			t.Fatal(got)
		}
		if len(tc.want) == 0 && got != nil {
			t.Fatal("empty result must be nil")
		}
		if len(got) > 0 {
			got[0] = "changed"
			if !slices.Equal(tc.in, before) {
				t.Fatal("shared result")
			}
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 고유 목록: [go map slice]
	// 원본: [go map go slice map]
}
