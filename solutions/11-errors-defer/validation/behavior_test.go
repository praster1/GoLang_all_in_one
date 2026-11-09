package main

import (
	"errors"
	"testing"
)

func TestValidation(t *testing.T) {
	for _, tc := range []struct {
		name            string
		age             int
		badName, badAge bool
	}{{"Go", 0, false, false}, {"Go", 150, false, false}, {" ", 0, true, false}, {"Go", -1, false, true}, {"", 151, true, true}} {
		err := validate(tc.name, tc.age)
		if errors.Is(err, ErrName) != tc.badName || errors.Is(err, ErrAge) != tc.badAge || (err == nil) != (!tc.badName && !tc.badAge) {
			t.Fatal(tc, err)
		}
	}
}
