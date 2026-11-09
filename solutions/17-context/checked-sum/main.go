package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func sum(ctx context.Context, values []int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if len(values) > 1000 {
		return 0, errors.New("값은 1000개까지 허용합니다")
	}
	total := 0
	for _, value := range values {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if value < 0 || value > 1000 {
			return 0, errors.New("각 값은 0~1000이어야 합니다")
		}
		total += value
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return total, nil
}

func run() error {
	total, err := sum(context.Background(), []int{10, 20, 30})
	if err != nil {
		return err
	}
	fmt.Println("합계:", total)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = sum(ctx, []int{10, 20, 30})
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
