package main

import "testing"

func TestStack(t *testing.T) {
	var s Stack[int]
	if got, ok := s.Pop(); ok || got != 0 {
		t.Fatal(got, ok)
	}
	s.Push(1)
	s.Push(2)
	if s.Len() != 2 {
		t.Fatal(s.Len())
	}
	for _, want := range []int{2, 1} {
		got, ok := s.Pop()
		if !ok || got != want {
			t.Fatal(got, ok)
		}
	}
	if s.Len() != 0 {
		t.Fatal(s.Len())
	}
}
func TestStackReleasesSlot(t *testing.T) {
	value := 1
	s := Stack[*int]{}
	s.Push(&value)
	backing := s.items
	got, ok := s.Pop()
	if !ok || got != &value || backing[0] != nil {
		t.Fatal("removed pointer retained")
	}
}
