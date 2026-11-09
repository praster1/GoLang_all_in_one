package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results, err := batch.Run(ctx, []string{"first", "cancel", "last"}, batch.Config{Workers: 1, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) {
		if job.Index == 1 {
			cancel()
			return "", ctx.Err()
		}
		return job.Input, nil
	})
	for _, result := range results {
		fmt.Printf("%d: %s attempts=%d\n", result.Job.Index, result.State, result.Attempts)
	}
	fmt.Println("취소 오류:", errors.Is(err, context.Canceled))
}
