package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

func sumParts(values []int, workers int) (int, error) {
	if workers < 1 || workers > 8 || len(values) > 1000 {
		return 0, errors.New("작업자 1~8명과 값 1000개 이하가 필요합니다")
	}
	for _, value := range values {
		if value < 0 || value > 1000 {
			return 0, errors.New("값은 0~1000이어야 합니다")
		}
	}
	if len(values) == 0 {
		return 0, nil
	}
	if workers > len(values) {
		workers = len(values)
	}
	partials := make([]int, workers)
	var group sync.WaitGroup
	for worker := range workers {
		group.Go(func() {
			total := 0
			for index := worker; index < len(values); index += workers {
				total += values[index]
			}
			partials[worker] = total
		})
	}
	group.Wait()
	total := 0
	for _, partial := range partials {
		total += partial
	}
	return total, nil
}

func main() {
	total, err := sumParts([]int{10, 20, 30, 40, 50}, 2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("합계:", total)
}
