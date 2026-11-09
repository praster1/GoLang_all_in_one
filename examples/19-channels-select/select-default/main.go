package main

import "fmt"

func main() {
	var disabled <-chan int
	ready := make(chan int, 1)
	ready <- 7
	select {
	case value := <-disabled:
		fmt.Println("비활성:", value)
	case value := <-ready:
		fmt.Println("준비된 값:", value)
	default:
		fmt.Println("즉시 처리 없음")
	}
	select {
	case value := <-ready:
		fmt.Println("추가 값:", value)
	default:
		fmt.Println("즉시 처리 없음")
	}
}
