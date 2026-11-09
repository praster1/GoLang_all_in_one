package main

import (
	"fmt"
	"strings"
)

func main() {
	text := " \tGo  한글\nGo "
	fmt.Printf("양끝 정리: %q\n", strings.TrimSpace(text))
	fields := strings.Fields(text)
	fmt.Printf("단어: %q\n", fields)
	fmt.Println("연결:", strings.Join(fields, " | "))
	fmt.Printf("쉼표 분리: %q\n", strings.Split("go,,map", ","))
	fmt.Println("글의 바이트 위치:", strings.Index("Go한글", "글"))
	fmt.Println("접두사:", strings.HasPrefix("Go한글", "Go"))
}
