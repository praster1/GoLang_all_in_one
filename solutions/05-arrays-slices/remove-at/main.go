package main

import "fmt"

func removeAt(values []int, index int) ([]int, bool) {
	if index < 0 || index >= len(values) {
		return nil, false
	}
	result := make([]int, 0, len(values)-1)
	result = append(result, values[:index]...)
	result = append(result, values[index+1:]...)
	return result, true
}
func main() {
	values := []int{10, 20, 30, 40}
	result, ok := removeAt(values, 1)
	fmt.Println("삭제 성공:", ok)
	fmt.Println("결과:", result)
	fmt.Println("원본:", values)
}
