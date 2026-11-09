package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func withJobTimeout(timeout time.Duration, process batch.Processor) (batch.Processor, error) {
	if timeout < 0 || timeout > time.Minute || process == nil {
		return nil, errors.New("제한 시간은 0~1m이며 작업 함수가 필요합니다")
	}
	return func(parent context.Context, job batch.Job) (string, error) {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()
		if err := ctx.Err(); err != nil {
			return "", err
		}
		value, err := process(ctx, job)
		if err != nil {
			return "", err
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return value, nil
	}, nil
}
func main() {
	process, err := withJobTimeout(0, func(ctx context.Context, job batch.Job) (string, error) { return job.Input, ctx.Err() })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	results, err := batch.Run(context.Background(), []string{"go"}, batch.Config{Workers: 1, Attempts: 3}, process)
	fmt.Println("항목 상태:", results[0].State)
	fmt.Println("시도 횟수:", results[0].Attempts)
	fmt.Println("항목 마감 오류:", errors.Is(err, context.DeadlineExceeded))
}
