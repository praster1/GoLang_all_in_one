package main

import "fmt"

func added(value, amount int) int {
	value += amount
	return value
}

func addInPlace(value *int, amount int) bool {
	if value == nil {
		return false
	}
	*value += amount
	return true
}

func redirect(value *int) {
	replacement := 99
	value = &replacement
	*value = 100
}

func main() {
	number := 10
	result := added(number, 3)
	fmt.Println("값으로 전달:", number, result)
	ok := addInPlace(&number, 5)
	fmt.Println("주소로 접근:", number, ok)
	redirect(&number)
	fmt.Println("포인터 재대입 뒤:", number)
	fmt.Println("nil 전달:", addInPlace(nil, 5))
}
