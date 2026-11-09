package main

import (
	"fmt"
	"sync"
)

func main() {
	results := make([]string, 3)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() {
			results[index] = fmt.Sprintf("작업 %d", index+1)
		})
	}
	group.Wait()
	for _, result := range results {
		fmt.Println(result)
	}
}
