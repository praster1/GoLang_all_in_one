package main

import "fmt"

func cloneInts(values []int) []int {
	if values == nil {
		return nil
	}
	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}
func main() {
	source := []int{1, 2, 3}
	cloned := cloneInts(source)
	cloned[0] = 99
	fmt.Println("원본:", source)
	fmt.Println("독립 복사:", cloned)
	short := make([]int, 2)
	count := copy(short, source)
	fmt.Printf("짧은 목적지: %v, 복사 수: %d\n", short, count)
	shifted := []int{1, 2, 3, 4}
	copy(shifted[1:], shifted[:3])
	fmt.Println("겹치는 복사:", shifted)
}
