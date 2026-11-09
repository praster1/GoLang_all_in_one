package main

import "fmt"

func uniqueNames(names []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}
func main() {
	names := []string{"Mina", "Jin", "Mina", "Sol", "Jin"}
	fmt.Println("처음 등장 순서:", uniqueNames(names))
	fmt.Println("원본:", names)
}
