package lesson

import "testing"

func TestNew(t *testing.T) {
	value, ok := New(" Go ", 30)
	if !ok || value.Summary() != "Go: 30분" {
		t.Fatal(value, ok)
	}
	for _, minutes := range []int{1, 240} {
		if _, ok := New("Go", minutes); !ok {
			t.Fatal(minutes)
		}
	}
	for _, tc := range []struct {
		title   string
		minutes int
	}{{" ", 30}, {"Go", 0}, {"Go", 241}} {
		if _, ok := New(tc.title, tc.minutes); ok {
			t.Fatal(tc)
		}
	}
}
