package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type CreateInput struct {
	Title   string
	Minutes int
}

type FieldError struct{ Field string }

func (e *FieldError) Error() string { return "invalid field: " + e.Field }

func validate(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	length := utf8.RuneCountInString(input.Title)
	if length < 1 || length > 40 {
		return CreateInput{}, &FieldError{Field: "title"}
	}
	if input.Minutes < 1 || input.Minutes > 180 {
		return CreateInput{}, &FieldError{Field: "minutes"}
	}
	return input, nil
}

func main() {
	input, err := validate(CreateInput{Title: " Go ", Minutes: 25})
	if err != nil {
		fmt.Println("예상하지 못한 오류:", err)
		return
	}
	fmt.Printf("정규화: %s, %d분\n", input.Title, input.Minutes)
	_, err = validate(CreateInput{Title: "Go", Minutes: 0})
	var field *FieldError
	if errors.As(err, &field) {
		fmt.Println("필드 오류:", field.Field)
	}
}
