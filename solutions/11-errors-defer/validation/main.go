package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrName = errors.New("name required")
var ErrAge = errors.New("age must be 0..150")

func validate(name string, age int) error {
	var all []error
	if strings.TrimSpace(name) == "" {
		all = append(all, ErrName)
	}
	if age < 0 || age > 150 {
		all = append(all, ErrAge)
	}
	return errors.Join(all...)
}
func main() {
	err := validate(" ", 151)
	fmt.Println("이름:", errors.Is(err, ErrName))
	fmt.Println("나이:", errors.Is(err, ErrAge))
	fmt.Println("정상:", validate("Go", 0) == nil)
}
