package main

import (
	"errors"
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func main() {
	_, err := report.ParseRow([]string{"도서", "bad"}, 3)
	var detail *report.DataError
	if errors.As(err, &detail) {
		fmt.Println("레코드:", detail.Record)
		fmt.Println("필드:", detail.Field)
	}
	fmt.Println("입력 오류:", errors.Is(err, report.ErrInput))
}
