package main

import (
	"fmt"
	"slices"
)

func groupBy[T any, K comparable](values []T, key func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, value := range values {
		k := key(value)
		result[k] = append(result[k], value)
	}
	return result
}

type Lesson struct {
	Language string
	Minutes  int
}

func main() {
	groups := groupBy([]Lesson{{"Go", 20}, {"SQL", 30}, {"Go", 40}}, func(v Lesson) string { return v.Language })
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		fmt.Println(key, groups[key])
	}
}
