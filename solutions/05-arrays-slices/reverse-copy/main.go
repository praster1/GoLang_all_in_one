package main

import "fmt"

func reversed(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	result := make([]int, len(values))
	for i, value := range values {
		result[len(values)-1-i] = value
	}
	return result
}
func main() {
	values := []int{10, 20, 30, 40}
	fmt.Println("뒤집은 복사:", reversed(values))
	fmt.Println("원본:", values)
}
