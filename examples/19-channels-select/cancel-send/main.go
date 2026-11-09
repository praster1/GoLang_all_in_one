package main

import (
	"context"
	"errors"
	"fmt"
)

func sendNumbers(ctx context.Context, count int, out chan<- int) error {
	defer close(out)
	if count < 0 || count > 1000 {
		return errors.New("개수는 0~1000이어야 합니다")
	}
	for value := 1; value <= count; value++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- value:
		}
	}
	return ctx.Err()
}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	values := make(chan int)
	done := make(chan error, 1)
	go func() { done <- sendNumbers(ctx, 1000, values) }()
	first := []int{}
	for range 3 {
		first = append(first, <-values)
	}
	cancel()
	err := <-done
	fmt.Println("먼저 받은 값:", first)
	fmt.Println("전송 취소 확인:", errors.Is(err, context.Canceled))
}
