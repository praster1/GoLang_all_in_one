package main

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrOutOfRange = errors.New("value out of range")

type FieldError struct {
	Field string
	Value string
	Cause error
}

func (e *FieldError) Error() string { return fmt.Sprintf("%s=%q: %v", e.Field, e.Value, e.Cause) }
func (e *FieldError) Unwrap() error { return e.Cause }
func parseAge(text string) (int, error) {
	age, err := strconv.Atoi(text)
	if err != nil {
		return 0, &FieldError{Field: "age", Value: text, Cause: err}
	}
	if age < 0 || age > 150 {
		return 0, &FieldError{Field: "age", Value: text, Cause: ErrOutOfRange}
	}
	return age, nil
}
func main() {
	_, err := parseAge("151")
	err = fmt.Errorf("profile: %w", err)
	var fieldErr *FieldError
	if errors.As(err, &fieldErr) {
		fmt.Println("필드:", fieldErr.Field, fieldErr.Value)
	}
	fmt.Println("범위 오류:", errors.Is(err, ErrOutOfRange))
	_, err = parseAge("many")
	var numberErr *strconv.NumError
	fmt.Println("숫자 오류:", errors.As(err, &numberErr))
}
