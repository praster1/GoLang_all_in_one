package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func cleanTags(items []string) ([]string, bool) {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !utf8.ValidString(item) {
			return nil, false
		}
		tag := strings.ToLower(strings.TrimSpace(item))
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		result = append(result, tag)
	}
	return result, true
}
func main() {
	tags := []string{" Go ", "한글", "GO", "\t", " Map "}
	result, ok := cleanTags(tags)
	fmt.Printf("태그: %q, 성공: %t\n", result, ok)
	fmt.Printf("원본: %q\n", tags)
}
