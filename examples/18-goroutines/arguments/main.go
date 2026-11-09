package main

import (
	"fmt"
	"sync"
)

func main() {
	number := 7
	result := 0
	var group sync.WaitGroup
	group.Add(1)
	go func(value int) {
		defer group.Done()
		result = value * 2
	}(number)
	number = 100
	group.Wait()
	fmt.Println("원래 변수:", number)
	fmt.Println("전달된 값의 결과:", result)
}
