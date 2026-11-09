package main

import (
	"strconv"
	"testing"
)

func TestMapMethod(t *testing.T) {
	input := Box[int]{Value: 42}
	calls := 0
	got := input.Map(func(v int) string { calls++; return strconv.Itoa(v) })
	var typed Box[string] = got
	if typed.Value != "42" || calls != 1 || input.Value != 42 {
		t.Fatal(got, calls, input)
	}
	empty := Box[string]{}.Map(func(s string) int { return len(s) })
	if empty.Value != 0 {
		t.Fatal(empty)
	}
}
