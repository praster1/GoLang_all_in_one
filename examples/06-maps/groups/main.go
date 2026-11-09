package main

import "fmt"

func groupParity(values []int) map[string][]int {
	groups := make(map[string][]int)
	for _, value := range values {
		key := "odd"
		if value%2 == 0 {
			key = "even"
		}
		groups[key] = append(groups[key], value)
	}
	return groups
}
func main() {
	groups := groupParity([]int{1, 2, 4, 5, 0})
	fmt.Println("짝수:", groups["even"])
	fmt.Println("홀수:", groups["odd"])
	fmt.Println("없는 그룹 길이:", len(groups["unknown"]))
}
