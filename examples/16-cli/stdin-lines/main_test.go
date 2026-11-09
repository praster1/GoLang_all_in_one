package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
func TestSummarize(t *testing.T) {
	for _, c := range []struct{ input, want string }{
		{"", "학습 항목: 0\n"}, {"Go\n\n  \nTest", "학습 항목: 2\n"}, {"Go\r\n한글\r\n", "학습 항목: 2\n"},
	} {
		var out bytes.Buffer
		if err := summarize(strings.NewReader(c.input), &out); err != nil || out.String() != c.want {
			t.Fatal(out.String(), err)
		}
	}
	want := errors.New("I/O failure")
	var out bytes.Buffer
	if err := summarize(failedReader{want}, &out); !errors.Is(err, want) || out.Len() != 0 {
		t.Fatal(err, out.String())
	}
	if err := summarize(strings.NewReader("Go"), failedWriter{want}); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if err := summarize(strings.NewReader(strings.Repeat("x", 128*1024)), &out); err == nil {
		t.Fatal("oversized token accepted")
	}
}
