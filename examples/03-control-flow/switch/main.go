package main

import "fmt"

func main() {
	day := "토"

	switch day {
	case "월", "화", "수", "목", "금":
		fmt.Println("학습하는 날")
	case "토", "일":
		fmt.Println("쉬는 날")
	default:
		fmt.Println("요일 확인")
	}
}
