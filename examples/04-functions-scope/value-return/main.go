package main

import "fmt"

func addMinutes(current, extra int) int {
	current += extra
	return current
}

func main() {
	minutes := 30
	updated := addMinutes(minutes, 15)

	fmt.Printf("반환값: %d\n", updated)
	fmt.Printf("원래 값: %d\n", minutes)
	minutes = updated
	fmt.Printf("대입 후: %d\n", minutes)
}
