package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

var errInput = errors.New("invalid text or limit")

func prefixRunes(text string, limit int) (string, error) {
	if limit < 0 || !utf8.ValidString(text) {
		return "", errInput
	}
	runes := []rune(text)
	if limit > len(runes) {
		limit = len(runes)
	}
	return string(runes[:limit]), nil
}

func main() {
	result, err := prefixRunes("Go한글", 3)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
}
