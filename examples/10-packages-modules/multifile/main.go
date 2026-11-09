package main

import "fmt"

func main() {
	fmt.Println("총 시간:", totalMinutes([]int{20, 30, 10}))
	fmt.Println("빈 입력:", totalMinutes(nil))
}
