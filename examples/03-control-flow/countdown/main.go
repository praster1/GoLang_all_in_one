package main

import "fmt"

func main() {
	remaining := 3

	for remaining > 0 {
		fmt.Println(remaining)
		remaining--
	}
	fmt.Println("시작!")
}
