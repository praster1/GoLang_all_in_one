package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	for _, c := range []struct {
		args []string
		code int
		out  string
	}{
		{nil, 0, "안녕하세요, Go!\n"},
		{[]string{"-name", "learner", "-upper"}, 0, "안녕하세요, LEARNER!\n"},
		{[]string{"-upper=false", "-name", "Go"}, 0, "안녕하세요, Go!\n"},
		{[]string{"-h"}, 0, ""}, {[]string{"-unknown"}, 2, ""},
		{[]string{"-name", " "}, 2, ""}, {[]string{"file", "-upper"}, 2, ""},
		{[]string{"-upper", "false"}, 2, ""}, {[]string{"--", "-upper"}, 2, ""},
	} {
		var out, diag bytes.Buffer
		code := run(c.args, &out, &diag)
		if code != c.code || out.String() != c.out {
			t.Fatalf("%v => %d %q %q", c.args, code, out.String(), diag.String())
		}
		if c.code == 2 && diag.Len() == 0 {
			t.Fatal("missing diagnostic")
		}
		if len(c.args) == 1 && c.args[0] == "-h" && !strings.Contains(diag.String(), "Usage") {
			t.Fatal(diag.String())
		}
	}
	want := errors.New("write failure")
	var diag bytes.Buffer
	if run(nil, failedWriter{want}, &diag) != 1 {
		t.Fatal("writer error must fail")
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
