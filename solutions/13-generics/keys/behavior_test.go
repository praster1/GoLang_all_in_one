package main

import (
	"slices"
	"testing"
)

type ID int

func TestKeys(t *testing.T) {
	source := map[ID]string{3: "c", 1: "a"}
	got := sortedKeys(source)
	if !slices.Equal(got, []ID{1, 3}) || len(source) != 2 {
		t.Fatal(got, source)
	}
	if keys := sortedKeys(map[string]int(nil)); keys == nil || len(keys) != 0 {
		t.Fatal("empty nonnil")
	}
}
