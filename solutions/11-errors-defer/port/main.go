package main

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrPortRange = errors.New("port must be 1..65535")

func parsePort(text string) (int, error) {
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("parse port: %w", err)
	}
	if value < 1 || value > 65535 {
		return 0, fmt.Errorf("port %d: %w", value, ErrPortRange)
	}
	return value, nil
}
func main() {
	port, err := parsePort("8080")
	fmt.Println("포트:", port, "성공:", err == nil)
	_, err = parsePort("0")
	fmt.Println("범위:", errors.Is(err, ErrPortRange))
	_, err = parsePort("abc")
	var detail *strconv.NumError
	fmt.Println("형식:", errors.As(err, &detail))
}
