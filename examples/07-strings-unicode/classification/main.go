package main

import (
	"fmt"
	"unicode"
)

func classify(text string) (int, int, int, int) {
	letters, digits, spaces, others := 0, 0, 0, 0
	for _, value := range text {
		switch {
		case unicode.IsLetter(value):
			letters++
		case unicode.IsDigit(value):
			digits++
		case unicode.IsSpace(value):
			spaces++
		default:
			others++
		}
	}
	return letters, digits, spaces, others
}
func main() {
	letters, digits, spaces, others := classify("Go 한글 42!")
	fmt.Printf("글자: %d, 숫자: %d, 공백: %d, 기타: %d\n", letters, digits, spaces, others)
}
