package main

import "fmt"

type Integer interface{ ~int | ~int64 }

func sum[T Integer](values []T) T {
	var total T
	for _, value := range values {
		total += value
	}
	return total
}

type Cents int64

func main() {
	fmt.Println("int:", sum([]int{1, 2, 3}))
	total := sum([]Cents{100, 250})
	fmt.Printf("정의 타입: %d (%T)\n", total, total)
	fmt.Println("빈 입력:", sum([]int64(nil)))
}
