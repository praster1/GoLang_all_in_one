package main

import "fmt"

func first[T any](values []T) (T, bool) {
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	return values[0], true
}
func main() {
	number, ok := first([]int{7, 9})
	fmt.Println("정수:", number, ok)
	word, ok := first([]string{"Go", "학습"})
	fmt.Println("문자열:", word, ok)
	missing, ok := first[int](nil)
	fmt.Println("빈 입력:", missing, ok)
}
