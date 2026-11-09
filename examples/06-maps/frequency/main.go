package main

import (
	"fmt"
	"sort"
)

func frequency(words []string) map[string]int {
	counts := make(map[string]int)
	for _, word := range words {
		counts[word]++
	}
	return counts
}
func sortedKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func main() {
	counts := frequency([]string{"go", "map", "go", "slice", "map", "go"})
	for _, key := range sortedKeys(counts) {
		fmt.Printf("%s: %d\n", key, counts[key])
	}
}
