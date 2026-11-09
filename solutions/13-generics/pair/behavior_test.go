package main

import (
	"strconv"
	"testing"
)

func TestPair(t *testing.T) {
	input := Pair[int, string]{Left: 7, Right: "Go"}
	var swapped Pair[string, int] = input.Swap()
	if swapped.Left != "Go" || swapped.Right != 7 {
		t.Fatal(swapped)
	}
	calls := 0
	var mapped Pair[string, string] = input.MapLeft(func(v int) string { calls++; return strconv.Itoa(v) })
	if mapped.Left != "7" || mapped.Right != "Go" || calls != 1 || input.Left != 7 {
		t.Fatal(mapped, calls, input)
	}
}
