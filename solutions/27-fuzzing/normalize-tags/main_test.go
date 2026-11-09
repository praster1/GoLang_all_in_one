package main

import (
	"slices"
	"strings"
	"testing"
)

func TestTags(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []string
		bad  bool
	}{
		{"", nil, false}, {" , ", nil, false}, {"Go,api,GO", []string{"api", "go"}, false},
		{"go lang", nil, true}, {"한글", nil, true}, {"İ", nil, true},
	} {
		got, err := normalizeTags(tc.in)
		if (err != nil) != tc.bad || !slices.Equal(got, tc.want) {
			t.Fatalf("got=%v err=%v", got, err)
		}
	}
}

func FuzzTags(f *testing.F) {
	for _, seed := range []string{"", "Go,api,GO", ",,", "한글", "test-1"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 65536 {
			t.Skip()
		}
		got, err := normalizeTags(text)
		if err != nil {
			return
		}
		if !slices.IsSorted(got) {
			t.Fatal("unsorted")
		}
		for i, tag := range got {
			if tag == "" || tag != strings.ToLower(tag) {
				t.Fatal("unnormalized")
			}
			if i > 0 && got[i-1] == tag {
				t.Fatal("duplicate")
			}
		}
		again, err := normalizeTags(strings.Join(got, ","))
		if err != nil || !slices.Equal(got, again) {
			t.Fatal("not idempotent")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// [api go test-1]
}
