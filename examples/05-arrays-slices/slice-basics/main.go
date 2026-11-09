package main

import "fmt"

func main() {
	var pending []int
	empty := []int{}
	values := make([]int, 3, 5)
	values[1] = 7
	fmt.Printf("nil: %t, 길이: %d\n", pending == nil, len(pending))
	fmt.Printf("empty: %t, 길이: %d\n", empty == nil, len(empty))
	fmt.Printf("원소: %v, 길이: %d, 용량: %d\n", values, len(values), cap(values))
	values[2] = 9
	fmt.Println("변경:", values)
}
