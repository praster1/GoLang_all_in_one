package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func applyAll(ctx context.Context, values []int, transform func(context.Context, int) (int, error)) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if transform == nil || len(values) > 1000 {
		return nil, errors.New("변환 함수와 1000개 이하의 값이 필요합니다")
	}
	results := make([]int, 0, len(values))
	for _, value := range values {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err := transform(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("값 %d 변환: %w", value, err)
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func run() error {
	normal, err := applyAll(context.Background(), []int{1, 2, 3}, func(ctx context.Context, value int) (int, error) {
		return value * 2, ctx.Err()
	})
	if err != nil {
		return err
	}
	fmt.Println("정상 결과:", normal)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results, err := applyAll(ctx, []int{1, 2, 3}, func(ctx context.Context, value int) (int, error) {
		if value == 2 {
			cancel()
		}
		return value * 2, nil
	})
	fmt.Println("실패 결과 없음:", results == nil)
	fmt.Println("취소 확인:", errors.Is(err, context.Canceled))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
