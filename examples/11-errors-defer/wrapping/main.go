package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrEmpty = errors.New("name is empty")

func normalizeName(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ErrEmpty
	}
	return text, nil
}
func loadName(text string) (string, error) {
	value, err := normalizeName(text)
	if err != nil {
		return "", fmt.Errorf("load name: %w", err)
	}
	return value, nil
}
func main() {
	_, err := loadName(" ")
	fmt.Println("오류:", err)
	fmt.Println("직접 비교:", err == ErrEmpty)
	fmt.Println("원인 비교:", errors.Is(err, ErrEmpty))
	fmt.Println("새 오류 비교:", errors.Is(err, errors.New("name is empty")))
}
