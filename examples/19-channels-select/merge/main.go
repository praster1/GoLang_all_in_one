package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

func merge(ctx context.Context, inputs ...<-chan int) (<-chan int, <-chan struct{}) {
	out := make(chan int)
	done := make(chan struct{})
	var group sync.WaitGroup
	for _, input := range inputs {
		if input == nil {
			continue
		}
		group.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case value, ok := <-input:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- value:
					}
				}
			}
		})
	}
	go func() { group.Wait(); close(out); close(done) }()
	return out, done
}
func main() {
	left := make(chan int, 2)
	right := make(chan int, 2)
	left <- 1
	left <- 3
	close(left)
	right <- 2
	right <- 4
	close(right)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, done := merge(ctx, left, right)
	values := []int{}
	for value := range out {
		values = append(values, value)
	}
	<-done
	sort.Ints(values)
	fmt.Println("병합 결과:", values)
}
