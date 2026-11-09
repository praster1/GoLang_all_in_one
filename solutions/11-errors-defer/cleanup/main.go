package main

import (
	"errors"
	"fmt"
	"io"
)

// withResource takes ownership of r; operation receives the same resource.
func withResource(r io.Closer, operation func(io.Closer) error) (err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close: %w", closeErr))
		}
	}()
	if callErr := operation(r); callErr != nil {
		return fmt.Errorf("operation: %w", callErr)
	}
	return nil
}

type resource struct {
	closeErr error
	closes   int
}

func (r *resource) Close() error { r.closes++; return r.closeErr }
func main() {
	workErr := errors.New("work failed")
	closeErr := errors.New("close failed")
	r := &resource{closeErr: closeErr}
	err := withResource(r, func(io.Closer) error { return workErr })
	fmt.Println("처리 오류:", errors.Is(err, workErr))
	fmt.Println("정리 오류:", errors.Is(err, closeErr))
	fmt.Println("닫기 횟수:", r.closes)
}
