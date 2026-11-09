package main

import (
	"errors"
	"testing"
	"unicode/utf8"
)

func TestReverse(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"", ""}, {"Go", "oG"}, {"Go한글", "글한oG"}, {"a\u0301", "\u0301a"}} {
		got, err := reverseRunes(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("got=%q err=%v", got, err)
		}
	}
	if _, err := reverseRunes(string([]byte{0xff})); !errors.Is(err, errUTF8) {
		t.Fatal("invalid accepted")
	}
}

func FuzzReverse(f *testing.F) {
	for _, seed := range []string{"", "Go", "한글", "a\u0301", string([]byte{0xff})} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 65536 {
			t.Skip()
		}
		got, err := reverseRunes(input)
		if !utf8.ValidString(input) {
			if !errors.Is(err, errUTF8) {
				t.Fatal("invalid accepted")
			}
			return
		}
		if err != nil || !utf8.ValidString(got) {
			t.Fatalf("invalid result %q: %v", got, err)
		}
		if utf8.RuneCountInString(got) != utf8.RuneCountInString(input) {
			t.Fatal("lost rune")
		}
		again, err := reverseRunes(got)
		if err != nil || again != input {
			t.Fatalf("round trip: %q", again)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 글한oG
}
