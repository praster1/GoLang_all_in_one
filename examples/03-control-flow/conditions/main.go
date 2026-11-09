package main

import "fmt"

func main() {
	score := 72

	if score >= 80 {
		fmt.Println("우수")
	} else if score >= 60 {
		fmt.Println("통과")
	} else {
		fmt.Println("다시 연습")
	}
	fmt.Println("판정 완료")
}
