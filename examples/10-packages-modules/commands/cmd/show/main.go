package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/commands/internal/lesson"
)

func main() {
	value, ok := lesson.New(" Go ", 30)
	if !ok {
		fmt.Println("입력 오류")
		return
	}
	fmt.Println(value.Summary())
}
