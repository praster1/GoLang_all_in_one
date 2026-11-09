package main

import (
	"fmt"
	"unicode/utf8"
)

func prefixRunes(text string, count int) (string, bool) {
	if count < 0 || !utf8.ValidString(text) {
		return "", false
	}
	runes := []rune(text)
	if count >= len(runes) {
		return text, true
	}
	return string(runes[:count]), true
}
func main() {
	prefix, ok := prefixRunes("Go한글", 3)
	fmt.Printf("앞 3룬: %q, 성공: %t\n", prefix, ok)
	empty, ok := prefixRunes("한글", 0)
	fmt.Printf("앞 0룬: %q, 성공: %t\n", empty, ok)
}
