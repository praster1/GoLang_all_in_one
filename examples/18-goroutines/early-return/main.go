package main

import (
	"fmt"
	"sync"
)

func main() {
	values := []int{-1, 0, 2}
	results := make([]string, len(values))
	var group sync.WaitGroup
	for index, value := range values {
		group.Add(1)
		go func(index, value int) {
			defer group.Done()
			if value <= 0 {
				results[index] = "건너뜀"
				return
			}
			results[index] = "처리"
		}(index, value)
	}
	group.Wait()
	for index, result := range results {
		fmt.Printf("%d: %s\n", values[index], result)
	}
}
