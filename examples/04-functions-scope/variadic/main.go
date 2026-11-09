package main

import "fmt"

func sum(numbers ...int) int {
	total := 0
	for _, number := range numbers {
		total += number
	}
	return total
}

func main() {
	fmt.Printf("인자 없음: %d\n", sum())
	fmt.Printf("세 값의 합: %d\n", sum(10, 20, 30))
}
