package main

import (
	"fmt"
	"sync"
)

func send(out chan<- int) {
	defer close(out)
	for _, value := range []int{2, 4, 6} {
		out <- value
	}
}
func sum(in <-chan int) int {
	total := 0
	for value := range in {
		total += value
	}
	return total
}
func main() {
	values := make(chan int)
	var group sync.WaitGroup
	group.Go(func() { send(values) })
	total := sum(values)
	group.Wait()
	fmt.Println("합계:", total)
}
