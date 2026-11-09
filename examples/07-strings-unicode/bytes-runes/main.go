package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	text := "Go한글"
	fmt.Printf("바이트: %d, 룬: %d\n", len(text), utf8.RuneCountInString(text))
	fmt.Printf("바이트 값: %x\n", text)
	for index, value := range text {
		fmt.Printf("%d: %c (%U)\n", index, value, value)
	}
}
