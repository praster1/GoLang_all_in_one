package main

import (
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func main() {
	summary := report.Summary{Rows: 2, Total: 3, Groups: []report.Group{{Category: "z", Count: 1, Total: 1}, {Category: "a", Count: 1, Total: 2}}}
	if err := report.WriteCSV(os.Stdout, summary); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("원본 첫 분류:", summary.Groups[0].Category)
}
