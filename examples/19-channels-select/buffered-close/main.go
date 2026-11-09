package main

import "fmt"

func main() {
	values := make(chan int, 2)
	values <- 10
	values <- 20
	close(values)
	collected := []int{}
	for value := range values {
		collected = append(collected, value)
	}
	value, ok := <-values
	fmt.Println("남아 있던 값:", collected)
	fmt.Printf("닫힌 뒤 수신: value=%d ok=%t\n", value, ok)
	fmt.Printf("길이=%d 용량=%d\n", len(values), cap(values))
}
