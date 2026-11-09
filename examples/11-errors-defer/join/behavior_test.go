package main

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		title              string
		count              int
		titleBad, countBad bool
	}{{"Go", 1, false, false}, {"", 1, true, false}, {"Go", 0, false, true}, {"", 0, true, true}} {
		err := validate(tc.title, tc.count)
		if errors.Is(err, ErrTitle) != tc.titleBad || errors.Is(err, ErrCount) != tc.countBad || (err == nil) != (!tc.titleBad && !tc.countBad) {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
}
