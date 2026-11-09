package main

import "fmt"

type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(value T) { s.items = append(s.items, value) }
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	last := len(s.items) - 1
	value := s.items[last]
	var zero T
	s.items[last] = zero
	s.items = s.items[:last]
	return value, true
}
func (s *Stack[T]) Len() int { return len(s.items) }

type Popper[T any] interface{ Pop() (T, bool) }

var _ Popper[string] = (*Stack[string])(nil)

func main() {
	var stack Stack[string]
	stack.Push("첫째")
	stack.Push("둘째")
	for stack.Len() > 0 {
		value, ok := stack.Pop()
		fmt.Println("꺼내기:", value, ok)
	}
	value, ok := stack.Pop()
	fmt.Printf("빈 스택: %q %t\n", value, ok)
}
