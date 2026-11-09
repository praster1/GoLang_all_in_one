package main

import (
	"slices"
	"testing"
)

func TestCleanTags(t *testing.T) {
	for _, tc := range []struct {
		in, want []string
		ok       bool
	}{
		{nil, nil, true}, {[]string{" ", "\t"}, nil, true},
		{[]string{"GO", "go", " 한글\u3000"}, []string{"go", "한글"}, true},
		{[]string{"good", "\xff"}, nil, false},
	} {
		before := slices.Clone(tc.in)
		got, ok := cleanTags(tc.in)
		if ok != tc.ok || !slices.Equal(got, tc.want) || !slices.Equal(tc.in, before) {
			t.Fatalf("%q %t", got, ok)
		}
		if len(tc.want) == 0 && got != nil {
			t.Fatal("empty/failed result must be nil")
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
	// 태그: ["go" "한글" "map"], 성공: true
	// 원본: [" Go " "한글" "GO" "\t" " Map "]
}
