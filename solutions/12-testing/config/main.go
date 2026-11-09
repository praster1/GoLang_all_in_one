package main

import (
	"errors"
	"fmt"
	"strconv"
)

const limitKey = "GOLANG_BOOK12_LIMIT"

var errLimit = errors.New("limit must be 1..100 decimal digits")

func loadLimit(lookup func(string) (string, bool)) (int, error) {
	text, present := lookup(limitKey)
	if !present {
		return 20, nil
	}
	if text == "" {
		return 0, errLimit
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return 0, errLimit
		}
	}
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 || n > 100 {
		return 0, errLimit
	}
	return n, nil
}

func main() {
	n, err := loadLimit(func(key string) (string, bool) { return "30", true })
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("한도:", n)
}
