package main

import (
	"errors"
	"testing"
)

func TestName(t *testing.T) {
	if got, err := loadName(" Go "); got != "Go" || err != nil {
		t.Fatalf("%q %v", got, err)
	}
	got, err := loadName(" \n")
	if got != "" || !errors.Is(err, ErrEmpty) {
		t.Fatalf("%q %v", got, err)
	}
}
