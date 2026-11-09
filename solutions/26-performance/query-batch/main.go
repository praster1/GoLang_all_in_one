package main

import "fmt"

func hitsLinear(values, queries []int) int {
	hits := 0
	for _, query := range queries {
		for _, value := range values {
			if value == query {
				hits++
				break
			}
		}
	}
	return hits
}

func index(values []int) map[int]struct{} {
	result := make(map[int]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func hitsIndexed(set map[int]struct{}, queries []int) int {
	hits := 0
	for _, query := range queries {
		if _, ok := set[query]; ok {
			hits++
		}
	}
	return hits
}

func main() {
	values := []int{2, 4, 6}
	queries := []int{4, 5, 4}
	fmt.Println("선형:", hitsLinear(values, queries))
	fmt.Println("인덱스:", hitsIndexed(index(values), queries))
}
