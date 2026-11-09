package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func TestRunDefaultAndInput(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := run(context.Background(), nil, strings.NewReader(""), &out, &diagnostics); code != 0 {
		t.Fatal(code, diagnostics.String())
	}
	var report batch.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary != (batch.Summary{Total: 3, Succeeded: 3, Attempts: 4}) || report.Items[1].Attempts != 2 || report.Items[1].Output != "CHANNELS" {
		t.Fatal(report)
	}
	out.Reset()
	diagnostics.Reset()
	if code := run(context.Background(), []string{"-input", "-"}, strings.NewReader("ok\nfail:x\n\n"), &out, &diagnostics); code != 1 {
		t.Fatal(code)
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Succeeded != 1 || report.Summary.Failed != 2 || report.Summary.Attempts != 3 {
		t.Fatal(report)
	}
}
func TestRunUsageAndCancellation(t *testing.T) {
	for _, args := range [][]string{{"-workers", "0"}, {"-attempts", "6"}, {"-timeout", "0s"}, {"-retry-delay", "-1s"}, {"-bad"}, {"unexpected"}} {
		if code := run(context.Background(), args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); code != 2 {
			t.Fatal(args, code)
		}
	}
	if code := run(context.Background(), []string{"-h"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal(code)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	if code := run(ctx, nil, strings.NewReader(""), &out, &bytes.Buffer{}); code != 130 {
		t.Fatal(code)
	}
	var report batch.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.Canceled != 3 || report.Summary.Attempts != 0 {
		t.Fatal(report)
	}
}
func TestReadLinesLimitsAndFile(t *testing.T) {
	for _, input := range []string{strings.Repeat("x", batch.MaxItemBytes+1), strings.Repeat("x\n", batch.MaxItems+1)} {
		if got, err := readLines(strings.NewReader(input)); got != nil || err == nil {
			t.Fatal(len(got), err)
		}
	}
	if got, err := readLines(strings.NewReader(strings.Repeat("x", batch.MaxItemBytes) + "\r\n")); err != nil || len(got) != 1 {
		t.Fatal(len(got), err)
	}
	name := filepath.Join(t.TempDir(), "jobs.txt")
	if err := os.WriteFile(name, []byte("go\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if code := run(context.Background(), []string{"-input", name}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); code != 0 {
		t.Fatal(code)
	}
	if code := run(context.Background(), []string{"-input", name + "-missing"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}); code != 1 {
		t.Fatal(code)
	}
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestReadAndWriteErrors(t *testing.T) {
	if got, err := readLines(failReader{}); got != nil || err == nil {
		t.Fatal(got, err)
	}
	if code := run(context.Background(), nil, strings.NewReader(""), failWriter{}, &bytes.Buffer{}); code != 1 {
		t.Fatal(code)
	}
}
