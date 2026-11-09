package main

import "fmt"

func main() {
	var counts map[string]int
	value, ok := counts["Go"]
	fmt.Printf("nil 조회: %d, %t, 길이: %d\n", value, ok, len(counts))
	delete(counts, "Go")
	clear(counts)
	counts = make(map[string]int)
	counts["Go"] = 1
	fmt.Printf("생성 뒤: %d, nil: %t\n", counts["Go"], counts == nil)
	clear(counts)
	fmt.Printf("비운 뒤 길이: %d, nil: %t\n", len(counts), counts == nil)
}
