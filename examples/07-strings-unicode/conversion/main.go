package main

import "fmt"

func main() {
	text := "Go한글"
	bytes := []byte(text)
	bytes[0] = 'N'
	runes := []rune(text)
	runes[2] = '말'
	fmt.Println("원본:", text)
	fmt.Println("바이트 수정:", string(bytes))
	fmt.Println("룬 수정:", string(runes))
	fmt.Printf("두 길이: %d, %d\n", len(bytes), len(runes))
}
