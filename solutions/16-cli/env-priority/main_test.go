package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestNamePriority(t *testing.T) {
	for _, c := range []struct {
		args   []string
		env    string
		exists bool
		want   string
		code   int
	}{
		{nil, "", false, "Go", 0}, {nil, " Env ", true, "Env", 0},
		{[]string{"-name", "CLI"}, "Env", true, "CLI", 0},
		{[]string{"-name", "CLI"}, "", true, "CLI", 0},
		{[]string{"-name", "Go"}, "Env", true, "Go", 0},
		{nil, "", true, "", 2}, {nil, " ", true, "", 2},
		{[]string{"-name="}, "Env", true, "", 2},
		{[]string{"-name", " "}, "Env", true, "", 2},
		{[]string{"-h"}, "", true, "", 0}, {[]string{"extra"}, "", false, "", 2},
	} {
		lookup := func(key string) (string, bool) {
			if key != "STUDY_NAME" {
				t.Fatal(key)
			}
			return c.env, c.exists
		}
		var out, diag bytes.Buffer
		code := run(c.args, lookup, &out, &diag)
		want := ""
		if c.want != "" {
			want = "이름: " + c.want + "\n"
		}
		if code != c.code || out.String() != want {
			t.Fatal(c.args, c.env, code, out.String(), diag.String())
		}
	}
	var diag bytes.Buffer
	lookup := func(string) (string, bool) { return "", false }
	if run(nil, lookup, failedWriter{errors.New("fail")}, &diag) != 1 {
		t.Fatal("write failure")
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
