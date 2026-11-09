package main

import (
	"errors"
	"testing"
)

func TestOpenFailure(t *testing.T) {
	if err := processOne(""); !errors.Is(err, ErrName) {
		t.Fatal(err)
	}
}
