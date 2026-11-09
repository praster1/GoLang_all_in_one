package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestStudy(t *testing.T) {
	for _, c := range []struct {
		args  []string
		input string
		code  int
		out   string
	}{
		{nil, "Go\nTest\n", 0, "항목: 2\n합계: 50분\n"},
		{[]string{"-json", "-minutes", "30"}, "Go\nTest", 0, "{\"count\":2,\"total_minutes\":60}\n"},
		{[]string{"-json=false"}, "", 0, "항목: 0\n합계: 0분\n"},
		{[]string{"-minutes", "1"}, "Go", 0, "항목: 1\n합계: 1분\n"},
		{[]string{"-minutes", "120"}, strings.Repeat("x\n", 100), 0, "항목: 100\n합계: 12000분\n"},
		{nil, strings.Repeat("x", 4096), 0, "항목: 1\n합계: 25분\n"},
		{nil, strings.Repeat("x", 4097), 2, ""}, {nil, strings.Repeat("x\n", 101), 2, ""},
		{[]string{"-minutes", "0"}, "", 2, ""}, {[]string{"-minutes", "121"}, "", 2, ""},
		{[]string{"-minutes", "x"}, "", 2, ""}, {[]string{"-json", "false"}, "", 2, ""},
		{[]string{"-h"}, "", 0, ""}, {[]string{"extra"}, "", 2, ""},
	} {
		var out, diag bytes.Buffer
		code := run(c.args, strings.NewReader(c.input), &out, &diag)
		if code != c.code || out.String() != c.out {
			t.Fatalf("%v len=%d => %d %q %q", c.args, len(c.input), code, out.String(), diag.String())
		}
		if c.code != 0 && diag.Len() == 0 {
			t.Fatal("missing diagnostic")
		}
	}
	var out, diag bytes.Buffer
	want := errors.New("I/O fail")
	if run(nil, failedReader{want}, &out, &diag) != 1 || out.Len() != 0 {
		t.Fatal("read failure")
	}
	for _, args := range [][]string{nil, {"-json"}} {
		if run(args, strings.NewReader("Go"), failedWriter{want}, &diag) != 1 {
			t.Fatal("write failure")
		}
	}
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
