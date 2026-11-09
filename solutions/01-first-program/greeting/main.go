package main

import (
	"fmt"
	"strings"
)

func greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Go"
	}
	return fmt.Sprintf("안녕하세요, %s!", name)
}

func main() {
	fmt.Println(greeting("Go"))
}
