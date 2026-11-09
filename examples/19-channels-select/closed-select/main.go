package main

import (
	"fmt"
	"sort"
)

func main() {
	left := make(chan int, 1)
	right := make(chan int, 1)
	left <- 3
	right <- 1
	close(left)
	close(right)
	values := []int{}
	closed := 0
	for left != nil || right != nil {
		select {
		case value, ok := <-left:
			if !ok {
				left = nil
				closed++
				continue
			}
			values = append(values, value)
		case value, ok := <-right:
			if !ok {
				right = nil
				closed++
				continue
			}
			values = append(values, value)
		}
	}
	sort.Ints(values)
	fmt.Println("모은 값:", values)
	fmt.Println("닫힘 확인:", closed)
}
