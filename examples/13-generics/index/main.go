package main

import "fmt"

func index[T comparable](values []T, target T) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

type Code int

func main() {
	fmt.Println("정수:", index([]int{3, 7, 3}, 3))
	fmt.Println("문자열:", index([]string{"Go", "한글"}, "한글"))
	fmt.Println("없음:", index([]Code{1, 2}, Code(9)))
}
