package main

import (
	"errors"
	"fmt"
)

var ErrTitle = errors.New("title required")
var ErrCount = errors.New("count must be positive")

func validate(title string, count int) error {
	var problems []error
	if title == "" {
		problems = append(problems, ErrTitle)
	}
	if count <= 0 {
		problems = append(problems, ErrCount)
	}
	return errors.Join(problems...)
}
func main() {
	err := validate("", 0)
	fmt.Println("제목 오류:", errors.Is(err, ErrTitle))
	fmt.Println("개수 오류:", errors.Is(err, ErrCount))
	fmt.Println("정상:", validate("Go", 2) == nil)
}
