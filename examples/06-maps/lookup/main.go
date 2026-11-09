package main

import "fmt"

func main() {
	scores := map[string]int{"Go": 0}
	value, ok := scores["Go"]
	fmt.Printf("Go: %d, 존재: %t\n", value, ok)
	value, ok = scores["Rust"]
	fmt.Printf("Rust: %d, 존재: %t\n", value, ok)
	fmt.Println("조회 뒤 수:", len(scores))
	scores["Rust"]++
	fmt.Printf("추가한 Rust: %d, 등록 수: %d\n", scores["Rust"], len(scores))
}
