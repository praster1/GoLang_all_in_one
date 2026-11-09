package batch

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestReportContract(t *testing.T) {
	results := []Result{{Job: Job{Index: 0, Input: "a"}, Output: "A", State: Succeeded, Attempts: 2}, {Job: Job{Index: 1, Input: "b"}, State: Failed, Attempts: 1, Err: errors.New("bad")}, {Job: Job{Index: 2, Input: "c"}, State: Canceled}}
	var out bytes.Buffer
	if err := WriteJSON(&out, results); err != nil {
		t.Fatal(err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Summary != (Summary{Total: 3, Succeeded: 1, Failed: 1, Canceled: 1, Attempts: 3}) || report.Items[1].Error != "bad" || report.Items[0].Output != "A" {
		t.Fatal(report)
	}
	if err := WriteJSON(brokenWriter{}, results); err == nil {
		t.Fatal("ignored write failure")
	}
	report = NewReport(nil)
	if report.Items == nil || report.Summary.Total != 0 {
		t.Fatal(report)
	}
}
