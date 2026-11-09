package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPrefixRunes(t *testing.T) {
	for _, tc := range []struct {
		text string
		n    int
		want string
		bad  bool
	}{
		{"", 0, "", false}, {"Go한글", 3, "Go한", false}, {"한글", 10, "한글", false},
		{"Go", -1, "", true}, {string([]byte{0xff}), 0, "", true},
	} {
		got, err := prefixRunes(tc.text, tc.n)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("got=%q err=%v", got, err)
		}
	}
}

func FuzzPrefixRunes(f *testing.F) {
	f.Add("Go한글", 3)
	f.Add("", 0)
	f.Add("한", -1)
	f.Add(string([]byte{0xff}), 1)
	f.Fuzz(func(t *testing.T, text string, n int) {
		if len(text) > 65536 {
			t.Skip()
		}
		got, err := prefixRunes(text, n)
		if n < 0 || !utf8.ValidString(text) {
			if err == nil {
				t.Fatal("invalid accepted")
			}
			return
		}
		if err != nil || !utf8.ValidString(got) || !strings.HasPrefix(text, got) {
			t.Fatal("prefix contract")
		}
		wantCount := min(n, utf8.RuneCountInString(text))
		if utf8.RuneCountInString(got) != wantCount {
			t.Fatal("wrong length")
		}
		again, err := prefixRunes(got, n)
		if err != nil || again != got {
			t.Fatal("not idempotent")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// Go한
}
