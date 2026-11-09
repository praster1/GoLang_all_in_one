package main

import (
	"fmt"
	"slices"
)

func firstBytes(data []byte, n int) []byte {
	if n <= 0 || len(data) == 0 {
		return nil
	}
	if n > len(data) {
		n = len(data)
	}
	return slices.Clone(data[:n])
}

func main() {
	source := []byte("Go language")
	prefix := firstBytes(source, 2)
	source[0] = 'N'
	fmt.Println(string(prefix))
	fmt.Println(string(source))
}
