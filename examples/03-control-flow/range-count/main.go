package main

import "fmt"

func main() {
	count := 3

	for index := range count {
		fmt.Printf("%d번째 학습\n", index+1)
	}
}
