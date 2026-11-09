package main

import (
	"fmt"
	"strings"
)

const nameKey = "GOLANG_BOOK12_NAME"

func readName(getenv func(string) string) string {
	name := strings.TrimSpace(getenv(nameKey))
	if name == "" {
		return "방문자"
	}
	return name
}

func main() {
	name := readName(func(key string) string { return " 학습자 " })
	fmt.Println("이름:", name)
}
