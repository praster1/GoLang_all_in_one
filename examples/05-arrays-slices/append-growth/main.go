package main

import "fmt"

func main() {
	values := make([]int, 2, 4)
	values[0], values[1] = 10, 20
	before := values
	values = append(values, 30, 40)
	fmt.Println("추가 후:", values)
	fmt.Println("기존 뷰 확장:", before[:4])
	values = append(values, 50)
	values[0] = 99
	fmt.Println("새 저장 공간:", values)
	fmt.Println("기존 값:", before)
	fmt.Println("필요한 용량 확보:", cap(values) >= len(values))
}
