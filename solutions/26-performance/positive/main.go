package main

import "fmt"

func positivesGrow(values []int) []int {
	var result []int
	for _, value := range values {
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}

func positivesReserve(values []int) []int {
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}

func main() {
	values := []int{-2, 3, 0, 5, -1}
	fmt.Println(positivesGrow(values))
	fmt.Println(positivesReserve(values))
}
