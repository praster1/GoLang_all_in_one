package main

import (
	"fmt"
	"slices"
)

type Key interface{ ~int | ~string }

func sortedKeys[K Key, V any](values map[K]V) []K {
	result := make([]K, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	slices.Sort(result)
	return result
}
func main() {
	fmt.Println(sortedKeys(map[string]int{"b": 2, "a": 1}))
	fmt.Println(sortedKeys(map[int]string{10: "ten", 2: "two"}))
}
