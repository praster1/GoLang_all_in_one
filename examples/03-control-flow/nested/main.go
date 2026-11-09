package main

import "fmt"

func main() {
	for row := 1; row <= 2; row++ {
		for column := 1; column <= 3; column++ {
			fmt.Printf("(%d,%d)\n", row, column)
		}
	}
}
