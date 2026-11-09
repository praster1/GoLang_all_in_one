package main

import (
	"fmt"
	"sort"
)

func mergeCounts(a, b map[string]int) map[string]int {
	result := make(map[string]int)
	for key, value := range a {
		result[key] = value
	}
	for key, value := range b {
		result[key] += value
	}
	return result
}
func main() {
	first := map[string]int{"go": 2, "map": 1}
	second := map[string]int{"go": 3, "slice": 2}
	merged := mergeCounts(first, second)
	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, merged[key])
	}
	fmt.Println("원본 go:", first["go"], second["go"])
}
