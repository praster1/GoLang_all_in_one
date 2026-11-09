package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestExitCodes(t *testing.T) {
	for _, c := range []struct {
		args []string
		code int
		out  string
	}{
		{[]string{"0"}, 0, "학습일: 0\n"}, {[]string{"7"}, 0, "학습일: 7\n"},
		{nil, 2, ""}, {[]string{"-1"}, 2, ""}, {[]string{"8"}, 2, ""}, {[]string{"x"}, 2, ""}, {[]string{"1", "2"}, 2, ""},
	} {
		var out, diag bytes.Buffer
		code := run(c.args, &out, &diag)
		if code != c.code || out.String() != c.out {
			t.Fatal(c.args, code, out.String(), diag.String())
		}
		if c.code != 0 && diag.Len() == 0 {
			t.Fatal("missing diagnostic")
		}
	}
	var diag bytes.Buffer
	if run([]string{"1"}, failedWriter{errors.New("fail")}, &diag) != 1 {
		t.Fatal("write failure")
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
