package main

import (
	"fmt"
	"strconv"
)

// transform preserves element order; fn must be non-nil.
func transform[T, U any](values []T, fn func(T) U) []U {
	if values == nil {
		return nil
	}
	result := make([]U, len(values))
	for i, value := range values {
		result[i] = fn(value)
	}
	return result
}
func main() {
	labels := transform([]int{2, 4}, func(value int) string { return "#" + strconv.Itoa(value) })
	fmt.Println("변환:", labels)
	lengths := transform(labels, func(value string) int { return len(value) })
	fmt.Println("바이트 길이:", lengths)
}
