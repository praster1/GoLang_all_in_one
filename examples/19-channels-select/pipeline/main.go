package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
)

func numbers(ctx context.Context, group *sync.WaitGroup, count int) <-chan int {
	out := make(chan int)
	group.Go(func() {
		defer close(out)
		for value := 1; value <= count; value++ {
			select {
			case <-ctx.Done():
				return
			case out <- value:
			}
		}
	})
	return out
}
func squared(ctx context.Context, group *sync.WaitGroup, in <-chan int) <-chan int {
	out := make(chan int)
	group.Go(func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- value * value:
				}
			}
		}
	})
	return out
}
func firstSquares(count, take int) ([]int, error) {
	if count < 0 || count > 1000 || take < 0 || take > count {
		return nil, errors.New("0 <= take <= count <= 1000이 필요합니다")
	}
	ctx, cancel := context.WithCancel(context.Background())
	var group sync.WaitGroup
	defer func() { cancel(); group.Wait() }()
	out := squared(ctx, &group, numbers(ctx, &group, count))
	results := make([]int, 0, take)
	for range take {
		value, ok := <-out
		if !ok {
			return nil, errors.New("예상보다 일찍 종료되었습니다")
		}
		results = append(results, value)
	}
	return results, nil
}
func main() {
	values, err := firstSquares(1000, 3)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("앞의 제곱값:", values)
	fmt.Println("모든 단계 종료 확인")
}
