package main

import (
	"fmt"
	"sync"
)

func main() {
	values := make(chan int)
	var group sync.WaitGroup
	group.Go(func() { values <- 42 })
	value := <-values
	group.Wait()
	fmt.Println("전달한 값:", value)
	fmt.Println("전송 작업 종료")
}
