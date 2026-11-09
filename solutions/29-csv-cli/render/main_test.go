package main

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

type countedReader struct{ calls int }

func (r *countedReader) Read(p []byte) (int, error) { r.calls++; return 0, io.EOF }

func TestRender(t *testing.T) {
	input := "category,amount\nx,2\nx,3\n"
	csv, err := render(strings.NewReader(input), "csv")
	if err != nil || string(csv) != "category,count,total\nx,2,5\n" {
		t.Fatal("CSV result")
	}
	data, err := render(strings.NewReader(input), "json")
	if err != nil {
		t.Fatal(err)
	}
	var summary report.Summary
	if err := json.Unmarshal(data, &summary); err != nil || summary.Rows != 2 || summary.Total != 5 || len(summary.Groups) != 1 || summary.Groups[0] != (report.Group{Category: "x", Count: 2, Total: 5}) {
		t.Fatal("JSON result")
	}
	if data, err := render(strings.NewReader("category,amount\nx,bad\n"), "csv"); err == nil || data != nil {
		t.Fatal("partial output")
	}
	counted := &countedReader{}
	if _, err := render(counted, "xml"); !errors.Is(err, errFormat) || counted.calls != 0 {
		t.Fatal("format validation order")
	}
}

func Example_main() {
	main()
	// Output:
	// category,count,total
	// 도서,2,20
}
