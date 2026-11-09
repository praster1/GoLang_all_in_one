package main

import (
	"strings"
	"testing"
)

func TestJoinSeparated(t *testing.T) {
	for _, tc := range []struct {
		parts []string
		sep   string
	}{
		{nil, ","}, {[]string{""}, ","}, {[]string{"", ""}, ","},
		{[]string{"a", "b"}, ""}, {[]string{"가", "나"}, " / "},
	} {
		if got, want := joinSeparated(tc.parts, tc.sep), strings.Join(tc.parts, tc.sep); got != want {
			t.Fatalf("got=%q want=%q", got, want)
		}
	}
}

func BenchmarkJoinSeparated(b *testing.B) {
	for _, size := range []struct {
		name string
		n    int
	}{{"N10", 10}, {"N100", 100}} {
		b.Run(size.name, func(b *testing.B) {
			parts := make([]string, size.n)
			for i := range parts {
				parts[i] = "Go"
			}
			b.ReportAllocs()
			for b.Loop() {
				joinSeparated(parts, ",")
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 한글,Go,측정
}
