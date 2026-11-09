package main

import "fmt"

func printGreeting(name string) {
	fmt.Printf("안녕하세요, %s!\n", name)
}

func main() {
	printGreeting("Go")
	printGreeting("개발자")
}
