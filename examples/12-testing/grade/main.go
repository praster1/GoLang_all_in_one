package main

import (
	"errors"
	"fmt"
)

func grade(score int) (string, error) {
	if score < 0 || score > 100 {
		return "", errors.New("score must be 0..100")
	}
	switch {
	case score >= 80:
		return "우수", nil
	case score >= 60:
		return "통과", nil
	default:
		return "보충", nil
	}
}

func main() {
	result, err := grade(80)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("판정:", result)
}
