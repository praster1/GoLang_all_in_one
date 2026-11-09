package main

import (
	"errors"
	"strconv"
	"testing"
)

func TestPort(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int
		bad  bool
	}{{"1", 1, false}, {"65535", 65535, false}, {"0", 0, true}, {"65536", 0, true}, {"-1", 0, true}} {
		got, err := parsePort(tc.text)
		if got != tc.want || errors.Is(err, ErrPortRange) != tc.bad {
			t.Fatalf("%+v: %d %v", tc, got, err)
		}
	}
	for _, s := range []string{"", "a", " 80", "99999999999999999999999999999"} {
		got, err := parsePort(s)
		var detail *strconv.NumError
		if got != 0 || !errors.As(err, &detail) {
			t.Fatalf("%q: %d %v", s, got, err)
		}
	}
}
