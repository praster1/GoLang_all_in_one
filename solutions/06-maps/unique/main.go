package main

import "fmt"

func unique(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	return result
}
func main() {
	items := []string{"go", "map", "go", "slice", "map"}
	fmt.Println("고유 목록:", unique(items))
	fmt.Println("원본:", items)
}
