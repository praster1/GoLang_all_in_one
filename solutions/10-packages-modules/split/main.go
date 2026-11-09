package main

import "fmt"

func main() {
	average, ok := mean([]int{10, 20, 30})
	fmt.Println("평균:", average, ok)
	average, ok = mean(nil)
	fmt.Println("빈 입력:", average, ok)
}
