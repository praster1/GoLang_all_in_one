package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

func validName(text string) bool {
	if !utf8.ValidString(text) {
		return false
	}
	runes := []rune(text)
	if len(runes) < 2 || len(runes) > 8 {
		return false
	}
	if !unicode.IsLetter(runes[0]) {
		return false
	}
	for _, value := range runes[1:] {
		if !unicode.IsLetter(value) && !unicode.IsDigit(value) && value != '_' {
			return false
		}
	}
	return true
}
func main() {
	fmt.Println("한글_2:", validName("한글_2"))
	fmt.Println("2Go:", validName("2Go"))
	fmt.Println("Go 학습:", validName("Go 학습"))
}
