package main

import "fmt"

func replaceFirst(values [3]int) [3]int {
	values[0] = 99
	return values
}
func main() {
	values := [3]int{20, 30}
	fmt.Println("초기:", values)
	changed := replaceFirst(values)
	fmt.Println("원본:", values)
	fmt.Println("복사본:", changed)
	fmt.Println("길이:", len(values))
}
