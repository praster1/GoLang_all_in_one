package main

import (
	"errors"
	"testing"
)

func TestLookup(t *testing.T) {
	if got, err := lookup(map[string]int{"a": 0}, "a"); got != 0 || err != nil {
		t.Fatal(got, err)
	}
	for _, values := range []map[string]int{nil, {}, {"other": 1}} {
		got, err := lookup(values, "missing")
		var detail *LookupError
		if got != 0 || !errors.Is(err, ErrNotFound) || !errors.As(err, &detail) || detail.Key != "missing" {
			t.Fatalf("%d %v", got, err)
		}
	}
}
