package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errInterval = errors.New("invalid interval")

func parseInterval(text string) (int, int, error) {
	left, right, ok := strings.Cut(text, ":")
	if !ok || left == "" || right == "" {
		return 0, 0, errInterval
	}
	for _, part := range []string{left, right} {
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return 0, 0, errInterval
			}
		}
	}
	low, err := strconv.Atoi(left)
	if err != nil {
		return 0, 0, errInterval
	}
	high, err := strconv.Atoi(right)
	if err != nil || low > high || high > 10000 {
		return 0, 0, errInterval
	}
	return low, high, nil
}

func main() {
	low, high, err := parseInterval("2:10")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("범위: %d~%d\n", low, high)
}
