package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestLineCLI(t *testing.T) {
	for _, c := range []struct {
		args  []string
		input string
		code  int
		out   string
	}{
		{nil, "Go\n\n Test \r\n", 0, "항목: 2\n"}, {nil, "", 0, "항목: 0\n"},
		{[]string{"-limit", "0"}, " \n", 0, "항목: 0\n"}, {[]string{"-limit", "0"}, "Go", 2, ""},
		{[]string{"-limit", "2"}, "Go\nTest", 0, "항목: 2\n"},
		{[]string{"-limit", "1"}, "Go\nTest", 2, ""},
		{[]string{"-limit", "100"}, strings.Repeat("x\n", 100), 0, "항목: 100\n"},
		{[]string{"-limit", "100"}, strings.Repeat("x\n", 101), 2, ""},
		{[]string{"-limit", "-1"}, "", 2, ""}, {[]string{"-limit", "101"}, "", 2, ""},
		{[]string{"-limit", "x"}, "", 2, ""}, {[]string{"-h"}, "", 0, ""}, {[]string{"extra"}, "", 2, ""},
	} {
		var out, diag bytes.Buffer
		code := run(c.args, strings.NewReader(c.input), &out, &diag)
		if code != c.code || out.String() != c.out {
			t.Fatal(c.args, code, out.String(), diag.String())
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
