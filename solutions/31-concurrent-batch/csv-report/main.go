package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func writeCSV(out io.Writer, results []batch.Result) error {
	writer := csv.NewWriter(out)
	if err := writer.Write([]string{"index", "input", "state", "attempts", "output", "error"}); err != nil {
		return err
	}
	for _, result := range results {
		message := ""
		if result.Err != nil {
			message = result.Err.Error()
		}
		record := []string{strconv.Itoa(result.Job.Index), result.Job.Input, string(result.State), strconv.Itoa(result.Attempts), result.Output, message}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
func run() error {
	results, err := batch.Run(context.Background(), []string{"go", "한글"}, batch.Config{Workers: 2, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) { return strings.ToUpper(job.Input), ctx.Err() })
	if err != nil {
		return err
	}
	return writeCSV(os.Stdout, results)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
