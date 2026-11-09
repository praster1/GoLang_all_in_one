package main

import (
	"fmt"
	"strings"
)

func joinSeparated(parts []string, sep string) string {
	var result strings.Builder
	for i, part := range parts {
		if i > 0 {
			result.WriteString(sep)
		}
		result.WriteString(part)
	}
	return result.String()
}

func main() { fmt.Println(joinSeparated([]string{"한글", "Go", "측정"}, ",")) }
