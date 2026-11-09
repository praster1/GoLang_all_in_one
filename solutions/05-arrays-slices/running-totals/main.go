package main

import "fmt"

func runningTotals(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	result := make([]int, len(values))
	total := 0
	for i, value := range values {
		total += value
		result[i] = total
	}
	return result
}
func main() {
	minutes := []int{20, 0, 30, 15}
	fmt.Println("일별:", minutes)
	fmt.Println("누적:", runningTotals(minutes))
}
