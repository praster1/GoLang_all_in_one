package main

import "fmt"

func increment(counts map[string]int, key string) bool {
	if counts == nil {
		return false
	}
	counts[key]++
	return true
}
func cloneCounts(counts map[string]int) map[string]int {
	if counts == nil {
		return nil
	}
	cloned := make(map[string]int, len(counts))
	for key, value := range counts {
		cloned[key] = value
	}
	return cloned
}
func main() {
	counts := map[string]int{"Go": 2}
	ok := increment(counts, "Go")
	fmt.Printf("변경: %d, 성공: %t\n", counts["Go"], ok)
	cloned := cloneCounts(counts)
	cloned["Go"] = 99
	fmt.Println("원본:", counts["Go"])
	fmt.Println("복사:", cloned["Go"])
	fmt.Println("nil 변경:", increment(nil, "Go"))
}
