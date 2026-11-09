package main

import (
	"errors"
	"fmt"
	"strconv"
)

var errCount = errors.New("count must be 1..100 decimal digits")

type countError struct{ Input string }

func (e *countError) Error() string { return fmt.Sprintf("count %q: %v", e.Input, errCount) }
func (e *countError) Unwrap() error { return errCount }

func parseCount(text string) (int, error) {
	fail := func() (int, error) { return 0, &countError{Input: text} }
	if text == "" {
		return fail()
	}
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return fail()
		}
	}
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 || n > 100 {
		return fail()
	}
	return n, nil
}

func main() {
	n, err := parseCount("12")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("수량:", n)
}
