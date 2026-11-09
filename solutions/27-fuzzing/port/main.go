package main

import (
	"errors"
	"fmt"
	"strconv"
)

var errPort = errors.New("invalid port")

func parsePort(text string) (uint16, error) {
	if text == "" {
		return 0, errPort
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return 0, errPort
		}
	}
	value, err := strconv.ParseUint(text, 10, 16)
	if err != nil || value == 0 {
		return 0, errPort
	}
	return uint16(value), nil
}

func main() {
	port, err := parsePort("08080")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("포트:", port)
}
