package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type LookupError struct {
	Key   string
	Cause error
}

func (e *LookupError) Error() string { return fmt.Sprintf("lookup %q: %v", e.Key, e.Cause) }
func (e *LookupError) Unwrap() error { return e.Cause }
func lookup(values map[string]int, key string) (int, error) {
	value, ok := values[key]
	if !ok {
		return 0, &LookupError{Key: key, Cause: ErrNotFound}
	}
	return value, nil
}
func main() {
	values := map[string]int{"zero": 0}
	value, err := lookup(values, "zero")
	fmt.Println("존재:", value, err == nil)
	_, err = lookup(values, "missing")
	fmt.Println("없음:", errors.Is(err, ErrNotFound))
	var detail *LookupError
	if errors.As(err, &detail) {
		fmt.Println("키:", detail.Key)
	}
}
