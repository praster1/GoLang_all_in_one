package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
)

func positive(ctx context.Context, values []int) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(values) > 1000 {
		return nil, errors.New("값은 1000개까지 허용합니다")
	}
	in := make(chan int)
	out := make(chan int)
	var group sync.WaitGroup
	group.Go(func() {
		defer close(in)
		for _, value := range values {
			select {
			case <-ctx.Done():
				return
			case in <- value:
			}
		}
	})
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
				if value > 0 {
					select {
					case <-ctx.Done():
						return
					case out <- value:
					}
				}
			}
		}
	})
	results := []int{}
	for value := range out {
		results = append(results, value)
	}
	group.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
func main() {
	values, err := positive(context.Background(), []int{-1, 2, 0, 3})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("양수:", values)
}
