package main

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type brokenWriter struct{}

func (brokenWriter) Write(p []byte) (int, error) { return 0, errors.New("write failed") }

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{nil, 0, "안녕하세요, Go!\n"}, {[]string{"-name", "CI"}, 0, "안녕하세요, CI!\n"},
		{[]string{"-name", " "}, 2, ""}, {[]string{"extra"}, 2, ""}, {[]string{"-bad"}, 2, ""}, {[]string{"-h"}, 0, ""},
	} {
		var out, errOut bytes.Buffer
		if code := run(tc.args, &out, &errOut); code != tc.code || out.String() != tc.want {
			t.Fatalf("args=%v code=%d out=%q", tc.args, code, out.String())
		}
	}
	if run(nil, brokenWriter{}, io.Discard) != 1 {
		t.Fatal("writer failure")
	}
}
