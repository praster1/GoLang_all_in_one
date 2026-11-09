package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func topGroups(summary report.Summary, n int) ([]report.Group, error) {
	if n < 0 {
		return nil, errors.New("n must be non-negative")
	}
	groups := slices.Clone(summary.Groups)
	slices.SortFunc(groups, func(a, b report.Group) int {
		if a.Total > b.Total {
			return -1
		}
		if a.Total < b.Total {
			return 1
		}
		return strings.Compare(a.Category, b.Category)
	})
	if n > len(groups) {
		n = len(groups)
	}
	return groups[:n], nil
}

func main() {
	summary, err := report.Analyze(strings.NewReader("category,amount\n도서,12\n식품,20\n도서,8\n"), report.DefaultOptions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	groups, err := topGroups(summary, 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for i, group := range groups {
		fmt.Printf("%d. %s: %d\n", i+1, group.Category, group.Total)
	}
}
