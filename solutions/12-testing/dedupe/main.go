package main

import "fmt"

func unique(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func main() { fmt.Println(unique([]string{"Go", "한글", "Go", "go"})) }
