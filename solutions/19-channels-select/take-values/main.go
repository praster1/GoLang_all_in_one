package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func take(ctx context.Context, in <-chan int, limit int) ([]int, error) {
	if limit < 0 || limit > 1000 {
		return nil, errors.New("개수는 0~1000이어야 합니다")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	results := make([]int, 0, limit)
	for len(results) < limit {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case value, ok := <-in:
			if !ok {
				return results, nil
			}
			results = append(results, value)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
func main() {
	in := make(chan int, 4)
	for value := 1; value <= 4; value++ {
		in <- value
	}
	close(in)
	values, err := take(context.Background(), in, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("받은 값:", values)
	fmt.Println("남은 첫 값:", <-in)
}
