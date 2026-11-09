package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

func wordCounts(text string) (map[string]int, bool) {
	if !utf8.ValidString(text) {
		return nil, false
	}
	counts := make(map[string]int)
	for _, word := range strings.Fields(strings.ToLower(text)) {
		counts[word]++
	}
	return counts, true
}
func main() {
	counts, ok := wordCounts(" Go\t한글 GO\n한글 ")
	if !ok {
		fmt.Println("텍스트 확인")
		return
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%q: %d\n", key, counts[key])
	}
}
