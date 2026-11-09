package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Add(delta int) { c.mu.Lock(); defer c.mu.Unlock(); c.value += delta }
func (c *Counter) Value() int    { c.mu.Lock(); defer c.mu.Unlock(); return c.value }
func main() {
	var counter Counter
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			for range 100 {
				counter.Add(1)
			}
		})
	}
	group.Wait()
	fmt.Println("합계:", counter.Value())
}
