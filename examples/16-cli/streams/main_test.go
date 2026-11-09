package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestStreams(t *testing.T) {
	var out, diag bytes.Buffer
	if err := writeResult(&out, &diag); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Go\nTest\n" || diag.String() != "처리: 2개\n" {
		t.Fatal(out.String(), diag.String())
	}
	want := errors.New("write failure")
	if err := writeResult(failedWriter{want}, &diag); !errors.Is(err, want) {
		t.Fatal(err)
	}
	out.Reset()
	if err := writeResult(&out, failedWriter{want}); !errors.Is(err, want) || out.Len() != 0 {
		t.Fatal(err, out.String())
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
