package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func count(workers, times int) int64 {
	var total atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < times; j++ {
				total.Add(1)
			}
		}()
	}
	wg.Wait()
	return total.Load()
}

func main() { fmt.Println("처리:", count(4, 100)) }
