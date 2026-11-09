package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

func sequential(ctx context.Context, inputs []string) ([]string, error) {
	if len(inputs) > 1000 {
		return nil, errors.New("입력은 1000개까지 허용합니다")
	}
	results := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		value := strings.TrimSpace(input)
		if value == "" {
			return nil, errors.New("빈 항목")
		}
		results = append(results, strings.ToUpper(value))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
func main() {
	values, err := sequential(context.Background(), []string{"go", " test ", "한글"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("순차 결과:", values)
}
