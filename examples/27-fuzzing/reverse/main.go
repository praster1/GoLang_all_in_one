package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

var errUTF8 = errors.New("invalid UTF-8")

func reverseRunes(text string) (string, error) {
	if !utf8.ValidString(text) {
		return "", errUTF8
	}
	values := []rune(text)
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
	return string(values), nil
}

func main() {
	result, err := reverseRunes("Go한글")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
}
