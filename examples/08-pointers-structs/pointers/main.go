package main

import "fmt"

func main() {
	value := 10
	pointer := &value
	fmt.Printf("타입: %T, 읽은 값: %d\n", pointer, *pointer)
	*pointer += 5
	fmt.Println("원래 변수:", value)
	alias := pointer
	*alias = 20
	other := 20
	fmt.Println("같은 변수:", pointer == &value)
	fmt.Println("같은 숫자, 다른 변수:", pointer == &other)
	fmt.Println("두 경로:", value, *alias)
}
