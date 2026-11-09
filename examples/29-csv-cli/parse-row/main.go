package main

import (
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func main() {
	row, err := report.ParseRow([]string{" 도서 ", "12000"}, 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s: %d\n", row.Category, row.Amount)
}
