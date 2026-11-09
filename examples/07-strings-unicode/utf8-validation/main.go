package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := string([]byte{'A', 0xff, 'B'})
	fmt.Println("유효한 UTF-8:", utf8.ValidString(text))
	for index, value := range text {
		fmt.Printf("%d: %U\n", index, value)
	}
	value, width := utf8.DecodeRuneInString(text[1:])
	fmt.Printf("잘못된 바이트: %U, 크기: %d\n", value, width)
	value, width = utf8.DecodeRuneInString("\uFFFD")
	fmt.Printf("실제 대체 문자: %U, 크기: %d\n", value, width)
}
