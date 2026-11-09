package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestShowArgs(t *testing.T) {
	var out bytes.Buffer
	if err := showArgs([]string{"Go", "file input"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "인자 수: 2\n0: \"Go\"\n1: \"file input\"\n" {
		t.Fatal(out.String())
	}
	out.Reset()
	if err := showArgs(nil, &out); err != nil || out.String() != "인자 수: 0\n" {
		t.Fatal(out.String(), err)
	}
	want := errors.New("write failure")
	if err := showArgs(nil, failedWriter{want}); !errors.Is(err, want) {
		t.Fatal(err)
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
