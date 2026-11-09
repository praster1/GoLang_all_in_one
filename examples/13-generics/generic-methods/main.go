package main

import (
	"fmt"
	"strconv"
)

type Box[T any] struct{ Value T }

// Map has its own U parameter in addition to the receiver's T (Go 1.27+).
func (box Box[T]) Map[U any](convert func(T) U) Box[U] {
	return Box[U]{Value: convert(box.Value)}
}
func main() {
	number := Box[int]{Value: 7}
	text := number.Map(func(value int) string { return "#" + strconv.Itoa(value) })
	fmt.Printf("원본: %d (%T)\n", number.Value, number)
	fmt.Printf("변환: %s (%T)\n", text.Value, text)
}
