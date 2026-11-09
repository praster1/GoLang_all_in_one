package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

var errFormat = errors.New("format must be csv or json")

func render(input io.Reader, format string) ([]byte, error) {
	if format != "csv" && format != "json" {
		return nil, errFormat
	}
	summary, err := report.Analyze(input, report.DefaultOptions())
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if format == "csv" {
		err = report.WriteCSV(&output, summary)
	} else {
		err = report.WriteJSON(&output, summary)
	}
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func main() {
	result, err := render(strings.NewReader("category,amount\n도서,12\n도서,8\n"), "csv")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(string(result))
}
