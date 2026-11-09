package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

func ordered(order []int) ([]int, []int, error) {
	if len(order) > 16 {
		return nil, nil, errors.New("작업은 16개까지 허용합니다")
	}
	seen := make([]bool, len(order))
	for _, index := range order {
		if index < 0 || index >= len(order) || seen[index] {
			return nil, nil, errors.New("0부터 작업 수-1까지의 순열이 필요합니다")
		}
		seen[index] = true
	}
	gates := make([]chan struct{}, len(order))
	for index := range gates {
		gates[index] = make(chan struct{})
	}
	results := make([]int, len(order))
	completed := make(chan int)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() { <-gates[index]; results[index] = index * 10; completed <- index })
	}
	actual := make([]int, 0, len(order))
	for _, index := range order {
		close(gates[index])
		actual = append(actual, <-completed)
	}
	group.Wait()
	return results, actual, nil
}
func main() {
	values, completed, err := ordered([]int{2, 0, 1})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("완료 순서:", completed)
	fmt.Println("입력 순서 결과:", values)
}
