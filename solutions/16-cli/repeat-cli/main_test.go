package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRepeat(t *testing.T) {
	for _, c := range []struct {
		args []string
		code int
		out  string
	}{
		{nil, 0, "안녕하세요, Go!\n"}, {[]string{"-name", " 민수 ", "-times", "2"}, 0, "안녕하세요, 민수!\n안녕하세요, 민수!\n"},
		{[]string{"-times", "5"}, 0, strings.Repeat("안녕하세요, Go!\n", 5)},
		{[]string{"-times", "0"}, 2, ""}, {[]string{"-times", "6"}, 2, ""}, {[]string{"-times", "x"}, 2, ""},
		{[]string{"-name", " "}, 2, ""}, {[]string{"-unknown"}, 2, ""}, {[]string{"extra"}, 2, ""}, {[]string{"-h"}, 0, ""},
	} {
		var out, diag bytes.Buffer
		code := run(c.args, &out, &diag)
		if code != c.code || out.String() != c.out {
			t.Fatal(c.args, code, out.String(), diag.String())
		}
		if c.code == 2 && diag.Len() == 0 {
			t.Fatal("missing diagnostic")
		}
	}
	var diag bytes.Buffer
	if run(nil, failedWriter{errors.New("fail")}, &diag) != 1 {
		t.Fatal("write failure")
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
