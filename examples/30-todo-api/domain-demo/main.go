package main

import (
	"errors"
	"fmt"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func main() {
	title, err := todo.NormalizeTitle("  Go 복습  ")
	fmt.Println("제목:", title, "성공:", err == nil)
	_, err = todo.NormalizeTitle(" \t ")
	var validation *todo.ValidationError
	fmt.Println("입력 오류:", errors.As(err, &validation))
	if validation != nil {
		fmt.Println("필드:", validation.Field)
	}
}
