package main

import (
	"slices"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func TestSelectCategory(t *testing.T) {
	source := report.Summary{Rows: 3, Total: 25, Groups: []report.Group{{Category: "도서", Count: 2, Total: 20}, {Category: "식품", Count: 1, Total: 5}}}
	before := slices.Clone(source.Groups)
	for _, tc := range []struct {
		name  string
		rows  int
		total int64
	}{{"도서", 2, 20}, {"식품", 1, 5}, {"없음", 0, 0}} {
		got := selectCategory(source, tc.name)
		if got.Rows != tc.rows || got.Total != tc.total || got.Groups == nil {
			t.Fatalf("got=%+v", got)
		}
		if len(got.Groups) > 0 {
			got.Groups[0].Total = 999
		}
		if !slices.Equal(source.Groups, before) {
			t.Fatal("mutated input")
		}
	}
}

func Example_main() {
	main()
	// Output:
	// category,count,total
	// 도서,2,20
}
