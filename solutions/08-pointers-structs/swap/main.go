package main

import "fmt"

func swap(left, right *int) bool {
	if left == nil || right == nil {
		return false
	}
	*left, *right = *right, *left
	return true
}

func main() {
	left, right := 3, 8
	fmt.Println("교환:", swap(&left, &right))
	fmt.Println("값:", left, right)
	fmt.Println("nil:", swap(nil, &right))
}
