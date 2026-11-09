package main

import (
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func main() {
	if err := report.WriteJSON(os.Stdout, report.Summary{}); err != nil {
		fmt.Println("error:", err)
	}
}
