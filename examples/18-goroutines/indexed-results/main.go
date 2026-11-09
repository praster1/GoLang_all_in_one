package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

func squares(values []int) ([]int, error) {
	if len(values) > 16 {
		return nil, errors.New("값은 16개까지 허용합니다")
	}
	for _, value := range values {
		if value < -1000 || value > 1000 {
			return nil, errors.New("각 값은 -1000~1000이어야 합니다")
		}
	}
	results := make([]int, len(values))
	var group sync.WaitGroup
	for index, value := range values {
		group.Add(1)
		go func(index, value int) {
			defer group.Done()
			results[index] = value * value
		}(index, value)
	}
	group.Wait()
	return results, nil
}

func main() {
	results, err := squares([]int{3, 1, 2})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("입력 순서의 결과:", results)
}
