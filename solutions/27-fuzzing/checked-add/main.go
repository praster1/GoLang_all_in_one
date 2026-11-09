package main

import (
	"errors"
	"fmt"
)

var errOverflow = errors.New("integer overflow")

func addChecked(a, b int) (int, error) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if b > 0 && a > maxInt-b || b < 0 && a < minInt-b {
		return 0, errOverflow
	}
	return a + b, nil
}

func main() {
	maxInt := int(^uint(0) >> 1)
	_, err := addChecked(maxInt, 1)
	fmt.Println("범위 오류:", errors.Is(err, errOverflow))
	total, err := addChecked(10, 20)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("합계:", total)
}
