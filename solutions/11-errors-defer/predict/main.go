package main

import "fmt"

func calculate() (result int) {
	result = 3
	defer fmt.Println("저장된 결과:", result)
	defer func() { result += 2; fmt.Println("정리 중 결과:", result) }()
	return 10
}
func main() { fmt.Println("최종:", calculate()) }
