package main

import (
	"errors"
	"io"
	"testing"
)

func TestBothFailures(t *testing.T) {
	a, b := errors.New("a"), errors.New("b")
	for _, tc := range []struct{ work, close error }{{nil, nil}, {a, nil}, {nil, b}, {a, b}} {
		r := &resource{closeErr: tc.close}
		calls := 0
		err := withResource(r, func(got io.Closer) error {
			calls++
			if got != r {
				t.Fatal("different resource")
			}
			return tc.work
		})
		if calls != 1 || r.closes != 1 || errors.Is(err, a) != (tc.work != nil) || errors.Is(err, b) != (tc.close != nil) || (err == nil) != (tc.work == nil && tc.close == nil) {
			t.Fatalf("%v calls=%d closes=%d", err, calls, r.closes)
		}
	}
}
