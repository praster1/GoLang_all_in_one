package main

import "fmt"

func sumEven(numbers ...int) int {
	total := 0
	for _, number := range numbers {
		if number%2 == 0 {
			total += number
		}
	}
	return total
}

func main() {
	fmt.Println("짝수 합계:", sumEven(1, 2, 3, 4, 6))
}
