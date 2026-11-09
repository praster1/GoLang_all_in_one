package main

import "testing"

func TestSum(t *testing.T) {
	if sum([]int(nil)) != 0 || sum([]int{-2, 5}) != 3 {
		t.Fatal("integer sum")
	}
	var got Cents = sum([]Cents{100, 250})
	if got != 350 {
		t.Fatal(got)
	}
}
