package main

import (
	"errors"
	"fmt"
)

var ErrName = errors.New("resource name required")

type resource struct{ name string }

func openResource(name string) (*resource, error) {
	if name == "" {
		return nil, ErrName
	}
	fmt.Println("열기:", name)
	return &resource{name: name}, nil
}
func (r *resource) Close() error { fmt.Println("닫기:", r.name); return nil }
func processOne(name string) (err error) {
	input, err := openResource(name)
	if err != nil {
		return fmt.Errorf("open resource: %w", err)
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close resource: %w", closeErr))
		}
	}()
	fmt.Println("처리:", name)
	return nil
}
func main() {
	for _, name := range []string{"A", "B"} {
		if err := processOne(name); err != nil {
			fmt.Println("실패:", err)
			return
		}
	}
}
