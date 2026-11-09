package main

import "fmt"

func main() {
	score := 72

	if score < 0 || score > 100 {
		fmt.Println("점수 확인")
	} else if score >= 90 {
		fmt.Println("우수")
	} else if score >= 60 {
		fmt.Println("통과")
	} else {
		fmt.Println("다시 연습")
	}
}
