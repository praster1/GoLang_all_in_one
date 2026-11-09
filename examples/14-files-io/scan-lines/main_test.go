package main

import (
	"errors"
	"strings"
	"testing"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestCountNonEmpty(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0}, {"\n \t\n", 0}, {"Go", 1},
		{"Go\r\n한글\r\n", 2}, {"Go\n\n끝", 2},
	}
	for _, tt := range tests {
		got, err := countNonEmpty(strings.NewReader(tt.input))
		if err != nil || got != tt.want {
			t.Errorf("countNonEmpty(%q) = (%d, %v), want %d", tt.input, got, err, tt.want)
		}
	}
}

func TestCountNonEmptyErrors(t *testing.T) {
	failure := errors.New("input failed")
	if _, err := countNonEmpty(failingReader{err: failure}); !errors.Is(err, failure) {
		t.Fatalf("error = %v, want input failure", err)
	}
	if _, err := countNonEmpty(strings.NewReader(strings.Repeat("x", 128*1024))); err == nil {
		t.Fatal("oversized token accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 비어 있지 않은 줄: 2
}
