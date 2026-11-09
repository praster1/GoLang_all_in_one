package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func TestMergeInputs(t *testing.T) {
	got, err := mergeInputs([]io.Reader{strings.NewReader("category,amount\nx,1\n"), strings.NewReader("category,amount\nx,2\ny,3\n")})
	if err != nil || got.Rows != 3 || got.Total != 6 || len(got.Groups) != 2 || got.Groups[0].Count != 2 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	got, err = mergeInputs(nil)
	if err != nil || got.Rows != 0 || got.Groups == nil {
		t.Fatal("empty inputs")
	}
	got, err = mergeInputs([]io.Reader{strings.NewReader("category,amount\nx,1\n"), strings.NewReader("category,amount\ny,bad\n")})
	if !errors.Is(err, report.ErrInput) || got.Rows != 0 || got.Total != 0 {
		t.Fatal("partial result")
	}
	inputs := make([]io.Reader, 10)
	for i := range inputs {
		inputs[i] = strings.NewReader("category,amount\n")
	}
	inputs[9] = strings.NewReader("category,amount\nx,7\n")
	got, err = mergeInputs(inputs)
	if err != nil || got.Rows != 1 || got.Total != 7 || len(got.Groups) != 1 {
		t.Fatalf("ten inputs: got=%+v err=%v", got, err)
	}
	if _, err := mergeInputs(make([]io.Reader, 11)); err == nil {
		t.Fatal("too many inputs")
	}
	if _, err := mergeInputs([]io.Reader{nil}); err == nil {
		t.Fatal("nil input")
	}
}

func TestMergeMaximumRowsAndAmounts(t *testing.T) {
	text := "category,amount\n" + strings.Repeat("x,1000000000\n", report.DefaultOptions().MaxRows)
	inputs := make([]io.Reader, 10)
	for i := range inputs {
		inputs[i] = strings.NewReader(text)
	}
	got, err := mergeInputs(inputs)
	if err != nil {
		t.Fatal(err)
	}
	const wantRows = 100000
	const wantTotal int64 = 100000000000000
	if got.Rows != wantRows || got.Total != wantTotal || len(got.Groups) != 1 ||
		got.Groups[0] != (report.Group{Category: "x", Count: wantRows, Total: wantTotal}) {
		t.Fatalf("maximum supported input: %+v", got)
	}
}

func Example_main() {
	main()
	// Output:
	// category,count,total
	// 도서,2,20
	// 식품,1,7
}
