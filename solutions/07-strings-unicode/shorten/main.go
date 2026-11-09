package main

import (
	"fmt"
	"unicode/utf8"
)

func shorten(text string, limit int) (string, bool) {
	if limit < 0 || !utf8.ValidString(text) {
		return "", false
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text, true
	}
	return string(runes[:limit]) + "…", true
}
func main() {
	result, ok := shorten("안녕하세요 Go", 5)
	fmt.Printf("요약: %q, 성공: %t\n", result, ok)
}
