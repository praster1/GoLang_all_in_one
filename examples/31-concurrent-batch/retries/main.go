package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	calls := make([]int, 3)
	results, err := batch.Run(context.Background(), []string{"ok", "temporary", "permanent"}, batch.Config{Workers: 2, Attempts: 3}, func(ctx context.Context, job batch.Job) (string, error) {
		calls[job.Index]++
		if job.Input == "temporary" && calls[job.Index] < 2 {
			return "", fmt.Errorf("busy: %w", batch.ErrTemporary)
		}
		if job.Input == "permanent" {
			return "", errors.New("bad input")
		}
		return strings.ToUpper(job.Input), ctx.Err()
	})
	for _, result := range results {
		fmt.Printf("%d: %s attempts=%d\n", result.Job.Index, result.State, result.Attempts)
	}
	var failure *batch.BatchError
	fmt.Println("항목 실패 보고:", errors.As(err, &failure))
}
