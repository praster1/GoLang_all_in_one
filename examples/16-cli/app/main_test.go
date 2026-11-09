package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestApp(t *testing.T) {
	for _, c := range []struct {
		args  []string
		input string
		code  int
		out   string
	}{
		{nil, "Go\n\nTest\n", 0, "학습 1: Go\n학습 2: Test\n"},
		{nil, "", 0, ""}, {[]string{"-limit", "1", "-prefix", "항목"}, " Go ", 0, "항목 1: Go\n"},
		{[]string{"-limit", "1"}, "Go\nTest", 2, ""}, {[]string{"-limit", "0"}, "", 2, ""},
		{[]string{"-limit", "101"}, "", 2, ""}, {[]string{"-prefix", " "}, "", 2, ""},
		{[]string{"-h"}, "", 0, ""}, {[]string{"file"}, "", 2, ""},
	} {
		var out, diag bytes.Buffer
		code := run(c.args, strings.NewReader(c.input), &out, &diag)
		if code != c.code || out.String() != c.out {
			t.Fatalf("%v %q => %d %q %q", c.args, c.input, code, out.String(), diag.String())
		}
	}
	var out, diag bytes.Buffer
	want := errors.New("I/O fail")
	if run(nil, failedReader{want}, &out, &diag) != 1 || out.Len() != 0 {
		t.Fatal("read failure")
	}
	if run(nil, strings.NewReader("Go"), failedWriter{want}, &diag) != 1 {
		t.Fatal("write failure")
	}
	if run(nil, strings.NewReader(strings.Repeat("x", 128*1024)), &out, &diag) != 1 {
		t.Fatal("token failure")
	}
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
