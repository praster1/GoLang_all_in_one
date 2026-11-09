package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

func runJoined(ctx context.Context, task func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if task == nil {
		return errors.New("작업 함수가 필요합니다")
	}
	var result error
	var group sync.WaitGroup
	group.Go(func() {
		result = task(ctx)
		if result == nil {
			result = ctx.Err()
		}
	})
	group.Wait()
	return result
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runJoined(ctx, func(ctx context.Context) error {
		cancel()
		<-ctx.Done()
		return ctx.Err()
	})
	fmt.Println("종료까지 기다림:", errors.Is(err, context.Canceled))
}
