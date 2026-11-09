package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type observed struct {
	io.Reader
	closeErr error
	calls    int
}

func (r *observed) Close() error { r.calls++; return r.closeErr }
func TestReadAndClose(t *testing.T) {
	readFailure := errors.New("read failure")
	closeFailure := errors.New("close failure")
	for _, tc := range []struct {
		name              string
		r                 io.Reader
		closeErr          error
		want              string
		readBad, closeBad bool
	}{{"success", strings.NewReader("Go"), nil, "Go", false, false}, {"read", failingReader{readFailure}, nil, "", true, false}, {"close", strings.NewReader("Go"), closeFailure, "", false, true}, {"both", failingReader{readFailure}, closeFailure, "", true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			r := &observed{Reader: tc.r, closeErr: tc.closeErr}
			got, err := readAndClose(r)
			if got != tc.want || r.calls != 1 || errors.Is(err, readFailure) != tc.readBad || errors.Is(err, closeFailure) != tc.closeBad || (err == nil) != (tc.name == "success") {
				t.Fatalf("%q err=%v calls=%d", got, err, r.calls)
			}
		})
	}
}
