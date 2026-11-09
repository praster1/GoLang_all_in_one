package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Add()       { c.mu.Lock(); defer c.mu.Unlock(); c.value++ }
func (c *Counter) Value() int { c.mu.Lock(); defer c.mu.Unlock(); return c.value }
func (c *Counter) Report(render func(int)) {
	snapshot := c.Value()
	// The callback can be slow or call Counter again, so invoke it after unlocking.
	render(snapshot)
}
func main() {
	var counter Counter
	counter.Add()
	counter.Report(func(value int) { counter.Add(); fmt.Println("보고서 기준 값:", value) })
	fmt.Println("현재 값:", counter.Value())
}
