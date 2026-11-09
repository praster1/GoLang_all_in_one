package main

import (
	"fmt"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func main() {
	input := "category,amount\n도서,12\n식품,5\n도서,8\n"
	summary, err := report.Analyze(strings.NewReader(input), report.DefaultOptions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("행: %d, 합계: %d, 분류: %d\n", summary.Rows, summary.Total, len(summary.Groups))
}
