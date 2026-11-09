package main

import (
	"fmt"
	"strconv"
)

type Pair[A, B any] struct {
	Left  A
	Right B
}

func (pair Pair[A, B]) Swap() Pair[B, A] { return Pair[B, A]{Left: pair.Right, Right: pair.Left} }
func (pair Pair[A, B]) MapLeft[C any](convert func(A) C) Pair[C, B] {
	return Pair[C, B]{Left: convert(pair.Left), Right: pair.Right}
}
func main() {
	pair := Pair[int, string]{Left: 7, Right: "Go"}
	fmt.Println("교환:", pair.Swap())
	mapped := pair.MapLeft(func(v int) string { return "#" + strconv.Itoa(v) })
	fmt.Printf("왼쪽 변환: %v (%T)\n", mapped, mapped)
	fmt.Println("원본:", pair)
}
