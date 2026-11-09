package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func mergeInputs(inputs []io.Reader) (report.Summary, error) {
	if len(inputs) > 10 {
		return report.Summary{}, errors.New("at most 10 inputs")
	}
	result := report.Summary{Groups: make([]report.Group, 0)}
	groups := make(map[string]report.Group)
	for i, input := range inputs {
		if input == nil {
			return report.Summary{}, fmt.Errorf("input %d: nil reader", i+1)
		}
		part, err := report.Analyze(input, report.DefaultOptions())
		if err != nil {
			return report.Summary{}, fmt.Errorf("input %d: %w", i+1, err)
		}
		result.Rows += part.Rows
		result.Total += part.Total
		for _, group := range part.Groups {
			current := groups[group.Category]
			current.Category = group.Category
			current.Count += group.Count
			current.Total += group.Total
			groups[group.Category] = current
		}
	}
	for _, group := range groups {
		result.Groups = append(result.Groups, group)
	}
	slices.SortFunc(result.Groups, func(a, b report.Group) int { return strings.Compare(a.Category, b.Category) })
	return result, nil
}

func main() {
	inputs := []io.Reader{strings.NewReader("category,amount\n도서,12\n"), strings.NewReader("category,amount\n식품,7\n도서,8\n")}
	summary, err := mergeInputs(inputs)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := report.WriteCSV(os.Stdout, summary); err != nil {
		fmt.Println("error:", err)
	}
}
