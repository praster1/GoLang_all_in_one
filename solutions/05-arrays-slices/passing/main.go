package main

import "fmt"

func passing(scores []int) []int {
	var result []int
	for _, score := range scores {
		if score >= 60 {
			result = append(result, score)
		}
	}
	return result
}
func main() {
	scores := []int{59, 60, 90, 30, 100}
	fmt.Println("통과 점수:", passing(scores))
	fmt.Println("원본:", scores)
}
