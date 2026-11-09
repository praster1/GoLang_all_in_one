package main

import (
	"errors"
	"fmt"
)

var errBounds = errors.New("lower bound exceeds upper bound")

func clamp(value, low, high int) (int, error) {
	if low > high {
		return 0, errBounds
	}
	if value < low {
		return low, nil
	}
	if value > high {
		return high, nil
	}
	return value, nil
}

func main() {
	value, err := clamp(120, 0, 100)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("보정:", value)
}
