package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var counter atomic.Int64
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			for range 100 {
				counter.Add(1)
			}
		})
	}
	group.Wait()
	fmt.Println("완료 수:", counter.Load())
}
