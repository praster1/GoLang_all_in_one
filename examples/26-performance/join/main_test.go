package main

import (
	"strings"
	"testing"
)

func TestJoin(t *testing.T) {
	for _, parts := range [][]string{nil, {}, {""}, {"가", "나"}, {"a", "", "b"}} {
		want := strings.Join(parts, "")
		if got := joinSlow(parts); got != want {
			t.Fatalf("slow=%q want=%q", got, want)
		}
		if got := joinBuilder(parts); got != want {
			t.Fatalf("builder=%q want=%q", got, want)
		}
	}
}

func BenchmarkJoin(b *testing.B) {
	parts := make([]string, 100)
	for i := range parts {
		parts[i] = "golang"
	}
	b.Run("Concat", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			joinSlow(parts)
		}
	})
	b.Run("Builder", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			joinBuilder(parts)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// Go로 측정
	// Go로 측정
}
