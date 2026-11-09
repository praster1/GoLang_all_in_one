package main

import "fmt"

func main() {
	var language string = "Go"
	completed := 1

	fmt.Printf("변경 전: %s %d장\n", language, completed)
	completed = completed + 1
	fmt.Printf("변경 후: %s %d장\n", language, completed)
}
