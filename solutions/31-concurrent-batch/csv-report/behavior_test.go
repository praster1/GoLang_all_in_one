package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"reflect"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestCSVQuotesAndErrors(t *testing.T) {
	var out bytes.Buffer
	results := []batch.Result{{Job: batch.Job{Index: 0, Input: "comma,quote\"\nline"}, State: batch.Failed, Attempts: 2, Err: errors.New("failed")}}
	if err := writeCSV(&out, results); err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(&out).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || !reflect.DeepEqual(records[1], []string{"0", "comma,quote\"\nline", "failed", "2", "", "failed"}) {
		t.Fatal(records)
	}
	if err := writeCSV(brokenWriter{}, results); err == nil {
		t.Fatal("write failure ignored")
	}
	out.Reset()
	if err := writeCSV(&out, nil); err != nil {
		t.Fatal(err)
	}
	records, err = csv.NewReader(&out).ReadAll()
	if err != nil || len(records) != 1 {
		t.Fatal(records, err)
	}
}
