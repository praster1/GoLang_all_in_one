package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/commands/internal/lesson"
)

func main() {
	_, ok := lesson.New("Go", 0)
	fmt.Println("유효:", ok)
}
