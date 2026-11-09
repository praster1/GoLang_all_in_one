package main

import "fmt"

func show() {
	value := 1
	defer fmt.Println("인자 저장:", value)
	defer func() { fmt.Println("나중에 읽기:", value) }()
	defer fmt.Println("마지막 등록")
	value = 2
	fmt.Println("본문:", value)
}
func main() { show(); fmt.Println("호출자") }
