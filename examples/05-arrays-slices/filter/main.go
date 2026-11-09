package main

import "fmt"

func atLeast(values []int, minimum int) []int {
	var selected []int
	for _, value := range values {
		if value >= minimum {
			selected = append(selected, value)
		}
	}
	return selected
}
func main() {
	minutes := []int{15, 40, 0, 25, 60}
	selected := atLeast(minutes, 30)
	fmt.Println("전체:", minutes)
	fmt.Println("30분 이상:", selected)
	selected[0] = 99
	fmt.Println("결과 수정 뒤 원본:", minutes)
}
