package main

import "fmt"

type Labeler interface{ Label() string }
type User struct{ Name string }

func (user *User) Label() string {
	if user == nil {
		return "사용자 없음"
	}
	return "사용자: " + user.Name
}

func main() {
	var empty Labeler
	var pointer *User
	var filled Labeler = pointer
	fmt.Println("빈 인터페이스:", empty == nil)
	fmt.Println("포인터:", pointer == nil)
	fmt.Println("포인터를 담은 인터페이스:", filled == nil)
	fmt.Printf("동적 타입: %T\n", filled)
	fmt.Println(filled.Label())
}
