package main

import "fmt"

func last[T any](values []T) (T, bool) {
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[len(values)-1], true
}
func main() {
	value, ok := last([]int{2, 4, 6})
	fmt.Println(value, ok)
	word, ok := last[string](nil)
	fmt.Printf("%q %t\n", word, ok)
}
