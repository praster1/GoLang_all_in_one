package main

import "fmt"

const courseName = "Go"

func printCourse() {
	fmt.Println("과정:", courseName)
}

func main() {
	score := 72
	message := "계속 연습"
	if score >= 60 {
		message := "통과"
		fmt.Println("블록 안:", message)
	}
	fmt.Println("블록 밖:", message)
	printCourse()
}
