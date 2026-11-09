package main

import "fmt"

func main() {
	signal := "yellow"

	switch signal {
	case "red":
		fmt.Println("정지")
	case "yellow":
		fmt.Println("주의")
	case "green":
		fmt.Println("진행")
	default:
		fmt.Println("신호 확인")
	}
}
