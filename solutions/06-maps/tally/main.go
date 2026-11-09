package main

import (
	"fmt"
	"sort"
)

func tally(items []string) map[string]int {
	counts := make(map[string]int)
	for _, item := range items {
		counts[item]++
	}
	return counts
}
func main() {
	counts := tally([]string{"pen", "book", "pen", "note", "book", "pen"})
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, counts[key])
	}
}
