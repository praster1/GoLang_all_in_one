package main

import (
	"context"
	"errors"
	"fmt"
)

func process(ctx context.Context, items []string, step func(string) error) (int, error) {
	completed := 0
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return completed, err
		}
		if err := step(item); err != nil {
			return completed, fmt.Errorf("항목 %q: %w", item, err)
		}
		completed++
	}
	return completed, ctx.Err()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	completed, err := process(ctx, []string{"Go", "Test", "Code"}, func(item string) error {
		if item == "Go" {
			cancel()
		}
		return nil
	})
	fmt.Println("완료 항목:", completed)
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
}
