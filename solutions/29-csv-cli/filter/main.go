package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func selectCategory(summary report.Summary, category string) report.Summary {
	result := report.Summary{Groups: make([]report.Group, 0)}
	for _, group := range summary.Groups {
		if group.Category == category {
			result.Groups = append(result.Groups, group)
			result.Rows += group.Count
			result.Total += group.Total
		}
	}
	return result
}

func main() {
	summary, err := report.Analyze(strings.NewReader("category,amount\n도서,12\n식품,5\n도서,8\n"), report.DefaultOptions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := report.WriteCSV(os.Stdout, selectCategory(summary, "도서")); err != nil {
		fmt.Println("error:", err)
	}
}
