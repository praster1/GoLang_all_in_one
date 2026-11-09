package main

import "fmt"

func contains(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func makeSet(values []int) map[int]struct{} {
	result := make(map[int]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func containsSet(set map[int]struct{}, target int) bool {
	_, found := set[target]
	return found
}

func main() {
	values := []int{3, 5, 8}
	set := makeSet(values)
	_, found := set[8]
	fmt.Println(contains(values, 8), found)
	_, found = set[7]
	fmt.Println(contains(values, 7), found)
}
