package main

import "fmt"

func main() {
	for number := 1; number <= 10; number++ {
		if number == 8 {
			break
		}
		if number%2 == 0 {
			continue
		}
		fmt.Println(number)
	}
	fmt.Println("반복 종료")
}
