package main

import (
	"errors"
	"fmt"
)

var errBounds = errors.New("minimum must not exceed maximum")

func clamp(value, minimum, maximum int) (int, error) {
	if minimum > maximum {
		return 0, errBounds
	}
	if value < minimum {
		return minimum, nil
	}
	if value > maximum {
		return maximum, nil
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
