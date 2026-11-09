package main

import (
	"bytes"
	"errors"
	"fmt"
)

var errLength = errors.New("invalid prefix length")

func clonePrefix(data []byte, n int) ([]byte, error) {
	if n < 0 || n > len(data) {
		return nil, errLength
	}
	if n == 0 {
		return nil, nil
	}
	return bytes.Clone(data[:n]), nil
}

func main() {
	source := []byte("Go!")
	result, err := clonePrefix(source, 2)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	source[0] = 'N'
	fmt.Printf("prefix=%s\nsource=%s\n", result, source)
}
