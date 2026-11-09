package main

import (
	"testing"
)

func TestCourseName(t *testing.T) {
	for _, c := range []struct {
		value  string
		exists bool
		want   string
		bad    bool
	}{
		{"", false, "Go", false}, {"Go", true, "Go", false}, {" Rust ", true, "Rust", false},
		{"", true, "", true}, {"  ", true, "", true},
	} {
		lookup := func(key string) (string, bool) {
			if key != "STUDY_COURSE" {
				t.Fatal(key)
			}
			return c.value, c.exists
		}
		got, err := courseName(lookup)
		if got != c.want || (err != nil) != c.bad {
			t.Fatalf("%q %t => %q %v", c.value, c.exists, got, err)
		}
	}
}
