package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

func TestWriteNumbered(t *testing.T) {
	cases := []struct{ input, want string }{
		{"", ""}, {"Go", "1: Go\n"}, {"Go\n", "1: Go\n"},
		{"Go\n\n끝", "1: Go\n2: \n3: 끝\n"},
		{"가\r\n나\r\n", "1: 가\n2: 나\n"},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		err := writeNumbered(&out, strings.NewReader(tc.input))
		if err != nil || out.String() != tc.want {
			t.Errorf("input %q: got %q, %v; want %q", tc.input, out.String(), err, tc.want)
		}
	}
}
func TestWriteNumberedPropagatesErrors(t *testing.T) {
	failure := errors.New("I/O failed")
	if err := writeNumbered(failedWriter{failure}, strings.NewReader("Go")); !errors.Is(err, failure) {
		t.Fatalf("writer error = %v", err)
	}
	var out bytes.Buffer
	if err := writeNumbered(&out, failedReader{failure}); !errors.Is(err, failure) {
		t.Fatalf("reader error = %v", err)
	}
	if err := writeNumbered(&out, strings.NewReader(strings.Repeat("x", 128*1024))); err == nil {
		t.Fatal("oversized token accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 1: Go
	// 2: 테스트
	// 3: 끝
}
