package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

func transformAll(ctx context.Context, inputs []string, transform func(context.Context, string) (string, error)) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(inputs) > 16 || transform == nil {
		return nil, errors.New("변환 함수와 입력 16개 이하가 필요합니다")
	}
	results := make([]string, len(inputs))
	failures := make([]error, len(inputs))
	var group sync.WaitGroup
	for index, input := range inputs {
		group.Go(func() {
			if err := ctx.Err(); err != nil {
				failures[index] = err
				return
			}
			value, err := transform(ctx, input)
			if err == nil {
				err = ctx.Err()
			}
			if err != nil {
				failures[index] = fmt.Errorf("항목 %d: %w", index, err)
				return
			}
			results[index] = value
		})
	}
	group.Wait()
	if err := errors.Join(failures...); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func run() error {
	results, err := transformAll(context.Background(), []string{"go", "test"}, func(ctx context.Context, input string) (string, error) {
		return strings.ToUpper(input), ctx.Err()
	})
	if err != nil {
		return err
	}
	fmt.Println("정상 결과:", results)
	results, err = transformAll(context.Background(), []string{"go", " "}, func(ctx context.Context, input string) (string, error) {
		if strings.TrimSpace(input) == "" {
			return "", errors.New("빈 입력")
		}
		return strings.ToUpper(input), ctx.Err()
	})
	fmt.Println("실패 결과 없음:", results == nil)
	fmt.Println("오류 확인:", err != nil)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
