package main

import (
	"fmt"
	"strconv"
)

func sumNumbers(values []string) (int, error) {
	total := 0
	for i, value := range values {
		number, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("item %d: %w", i+1, err)
		}
		total += number
	}
	return total, nil
}

func main() {
	total, err := sumNumbers([]string{"10", "20", "30"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("합계:", total)
}
