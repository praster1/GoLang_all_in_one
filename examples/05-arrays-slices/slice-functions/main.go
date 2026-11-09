package main

import "fmt"

func increment(values []int) {
	for i := range values {
		values[i]++
	}
}
func addValue(values []int, value int) []int {
	return append(values, value)
}
func main() {
	values := make([]int, 2, 4)
	values[0], values[1] = 10, 20
	increment(values)
	fmt.Println("원소 변경:", values)
	addValue(values, 30)
	fmt.Println("결과를 버린 길이:", len(values))
	values = addValue(values, 40)
	fmt.Println("반환값 대입:", values)
}
