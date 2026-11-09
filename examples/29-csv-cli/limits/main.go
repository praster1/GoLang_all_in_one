package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func main() {
	input := "category,amount\nx,1\n"
	options := report.Options{MaxBytes: int64(len(input)), MaxRows: 1}
	_, err := report.Analyze(strings.NewReader(input), options)
	fmt.Println("경계 허용:", err == nil)
	options.MaxBytes--
	_, err = report.Analyze(strings.NewReader(input), options)
	fmt.Println("초과 거부:", errors.Is(err, report.ErrLimit))
}
