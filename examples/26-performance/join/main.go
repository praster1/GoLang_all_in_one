package main

import (
	"fmt"
	"strings"
)

func joinSlow(parts []string) string {
	result := ""
	for _, part := range parts {
		result += part
	}
	return result
}

func joinBuilder(parts []string) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part)
	}
	return b.String()
}

func main() {
	parts := []string{"Go", "로", " ", "측정"}
	fmt.Println(joinSlow(parts))
	fmt.Println(joinBuilder(parts))
}
