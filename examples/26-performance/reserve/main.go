package main

import "fmt"

func squaresGrow(n int) []int {
	var result []int
	for i := 0; i < n; i++ {
		result = append(result, i*i)
	}
	return result
}

func squaresReserve(n int) []int {
	if n <= 0 {
		return nil
	}
	result := make([]int, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, i*i)
	}
	return result
}

func main() {
	fmt.Println(squaresGrow(5))
	fmt.Println(squaresReserve(5))
}
