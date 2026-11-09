package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestFrequency(t *testing.T) {
	got, err := countWords([][]string{{"Go", "test"}, {"go", "한글"}, {" test ", "GO"}})
	if err != nil || !reflect.DeepEqual(got, map[string]int{"go": 3, "test": 2, "한글": 1}) {
		t.Fatal(got, err)
	}
	got, err = countWords(nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	for _, input := range [][][]string{{{" "}}, {{string([]byte{0xff})}}, {{strings.Repeat("x", 65)}}, make([][]string, 9), {make([]string, 1001)}} {
		if got, err := countWords(input); got != nil || err == nil {
			t.Fatal(got, err)
		}
	}
}
