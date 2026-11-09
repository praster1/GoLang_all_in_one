package main

import (
	"errors"
	"strconv"
	"testing"
)

func TestAge(t *testing.T) {
	for _, age := range []string{"0", "150"} {
		if _, err := parseAge(age); err != nil {
			t.Fatal(err)
		}
	}
	for _, age := range []string{"-1", "151"} {
		got, err := parseAge(age)
		var field *FieldError
		if got != 0 || !errors.Is(err, ErrOutOfRange) || !errors.As(err, &field) || field.Field != "age" || field.Value != age {
			t.Fatalf("%d %v", got, err)
		}
	}
	_, err := parseAge("many")
	var n *strconv.NumError
	if !errors.As(err, &n) {
		t.Fatal(err)
	}
}
