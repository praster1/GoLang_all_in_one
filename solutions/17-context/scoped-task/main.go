package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

func withBudget(ctx context.Context, budget time.Duration, task func(context.Context) error) error {
	if budget < 0 || task == nil {
		return errors.New("음수 예산과 nil 작업은 허용하지 않습니다")
	}
	child, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	if err := child.Err(); err != nil {
		return err
	}
	if err := task(child); err != nil {
		return fmt.Errorf("작업: %w", err)
	}
	return child.Err()
}

func run() error {
	var observed context.Context
	err := withBudget(context.Background(), time.Hour, func(ctx context.Context) error {
		observed = ctx
		return ctx.Err()
	})
	if err != nil {
		return err
	}
	fmt.Println("작업 성공:", err == nil)
	fmt.Println("반환 후 정리:", errors.Is(observed.Err(), context.Canceled))
	called := false
	err = withBudget(context.Background(), 0, func(context.Context) error {
		called = true
		return nil
	})
	fmt.Println("0초 마감:", errors.Is(err, context.DeadlineExceeded))
	fmt.Println("0초 작업 실행:", called)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
