package main

import "fmt"

func newCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func main() {
	first := newCounter()
	second := newCounter()

	fmt.Println("첫 번째:", first())
	fmt.Println("첫 번째:", first())
	fmt.Println("두 번째:", second())
	fmt.Println("첫 번째:", first())
}
