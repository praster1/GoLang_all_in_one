package main

import (
	"errors"
	"fmt"
)

func divide(total, count int) (int, error) {
	if count <= 0 {
		return 0, errors.New("count must be positive")
	}
	return total / count, nil
}
func main() {
	value, err := divide(12, 3)
	if err != nil {
		fmt.Println("실패:", err)
		return
	}
	fmt.Println("몫:", value)
	_, err = divide(12, 0)
	if err != nil {
		fmt.Println("실패:", err)
	}
}
