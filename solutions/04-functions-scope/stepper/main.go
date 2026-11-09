package main

import "fmt"

func newStepper(start, step int) func() int {
	current := start
	return func() int {
		current += step
		return current
	}
}

func main() {
	first := newStepper(10, 3)
	second := newStepper(100, 10)
	fmt.Println("첫 번째:", first())
	fmt.Println("첫 번째:", first())
	fmt.Println("두 번째:", second())
	fmt.Println("첫 번째:", first())
}
