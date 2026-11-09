package main

import "fmt"

func add(a, b int) int {
	return a + b
}

func multiply(a, b int) int {
	return a * b
}

func apply(a, b int, operation func(int, int) int) int {
	return operation(a, b)
}

func main() {
	operation := add
	fmt.Println("덧셈:", operation(3, 4))
	operation = multiply
	fmt.Println("곱셈:", apply(3, 4, operation))

	double := func(value int) int {
		return value * 2
	}
	fmt.Println("두 배:", double(5))
}
