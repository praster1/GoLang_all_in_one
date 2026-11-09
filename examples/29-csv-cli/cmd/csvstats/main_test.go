package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const data = "category,amount\n도서,12\n식품,5\n도서,8\n"

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		input string
		code  int
		want  string
	}{
		{"stdin", nil, data, 0, "category,count,total\n도서,2,20\n식품,1,5\n"},
		{"empty-data", nil, "category,amount\n", 0, "category,count,total\n"},
		{"invalid", nil, "category,amount\nx,1\ny,bad\n", 1, ""},
		{"format", []string{"-format", "xml"}, data, 2, ""},
		{"rows", []string{"-max-rows", "0"}, data, 2, ""},
		{"limit", []string{"-max-rows", "1"}, data, 1, ""},
		{"argument", []string{"unexpected"}, data, 2, ""},
		{"help", []string{"-h"}, "", 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(tc.args, strings.NewReader(tc.input), &out, &errOut)
			if code != tc.code || out.String() != tc.want {
				t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
			}
			if tc.code != 0 && errOut.Len() == 0 {
				t.Fatal("missing error report")
			}
		})
	}
}

func TestFileAndJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sales.csv")
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"-input", path, "-format", "json"}, strings.NewReader("ignored"), &out, &errOut); code != 0 {
		t.Fatal(errOut.String())
	}
	var result struct {
		Rows   int
		Total  int64
		Groups []struct{ Category string }
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Rows != 3 || result.Total != 25 || len(result.Groups) != 2 {
		t.Fatal("JSON result")
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("file not closed: %v", err)
	}
	out.Reset()
	errOut.Reset()
	if code := run([]string{"-input", path}, strings.NewReader(data), &out, &errOut); code != 1 || out.Len() != 0 {
		t.Fatal("missing file")
	}
}

type failedWriter struct{}

func (failedWriter) Write(p []byte) (int, error) { return 0, errors.New("output unavailable") }

func TestOutputFailure(t *testing.T) {
	if code := run(nil, strings.NewReader(data), failedWriter{}, io.Discard); code != 1 {
		t.Fatal("output error ignored")
	}
}
