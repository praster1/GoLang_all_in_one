package main

import "fmt"

func main() {
	limit := 10
	total := 0
	count := 0

	for number := 1; number <= limit; number++ {
		if number%3 != 0 {
			continue
		}
		total += number
		count++
	}

	fmt.Printf("3의 배수: %d개\n합계: %d\n", count, total)
}
