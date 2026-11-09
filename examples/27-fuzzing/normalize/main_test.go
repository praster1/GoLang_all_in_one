package main

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	for _, tc := range []struct{ in, want string }{{"", ""}, {" \t\n", ""}, {"가\u3000나", "가 나"}, {" a  b ", "a b"}} {
		if got := normalizeSpace(tc.in); got != tc.want {
			t.Fatalf("got=%q", got)
		}
	}
}

func FuzzNormalize(f *testing.F) {
	for _, seed := range []string{"", " ", "a\tb\n", "가\u3000나", string([]byte{0xff})} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 65536 {
			t.Skip()
		}
		got := normalizeSpace(input)
		if normalizeSpace(got) != got {
			t.Fatal("not idempotent")
		}
		if strings.HasPrefix(got, " ") || strings.HasSuffix(got, " ") || strings.Contains(got, "  ") {
			t.Fatalf("spacing: %q", got)
		}
		if len(strings.Fields(got)) != len(strings.Fields(input)) {
			t.Fatal("lost field")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// Go 공부 시작
}
