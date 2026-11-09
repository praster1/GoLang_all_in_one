package main

import "fmt"

func main() {
	limit := 5
	total := 0

	for number := 1; number <= limit; number++ {
		total += number
		fmt.Printf("%d까지의 합: %d\n", number, total)
	}
	fmt.Printf("최종 합: %d\n", total)
}
