package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
)

func collect(ctx context.Context, parts [][]int) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(parts) > 8 {
		return nil, errors.New("입력 묶음은 8개까지 허용합니다")
	}
	total := 0
	for _, part := range parts {
		if len(part) > 1000-total {
			return nil, errors.New("전체 값은 1000개까지 허용합니다")
		}
		total += len(part)
	}
	out := make(chan int)
	var group sync.WaitGroup
	for _, part := range parts {
		group.Go(func() {
			for _, value := range part {
				select {
				case <-ctx.Done():
					return
				case out <- value:
				}
			}
		})
	}
	done := make(chan struct{})
	go func() { group.Wait(); close(out); close(done) }()
	results := make([]int, 0, total)
	for value := range out {
		results = append(results, value)
	}
	<-done
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sort.Ints(results)
	return results, nil
}
func main() {
	values, err := collect(context.Background(), [][]int{{3, 1}, {2, 3}, nil})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("모은 값:", values)
}
