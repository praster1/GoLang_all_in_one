package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	results, err := batch.Run(context.Background(), []string{"c", "a", "한글"}, batch.Config{Workers: 2, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) { return strings.ToUpper(job.Input), ctx.Err() })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, result := range results {
		fmt.Printf("%d: %s (%s)\n", result.Job.Index, result.Output, result.State)
	}
}
