package main

import "fmt"

func main() {
	var count int
	var rate float64
	var ready bool
	var title string

	fmt.Printf("count=%d (%T)\n", count, count)
	fmt.Printf("rate=%.1f (%T)\n", rate, rate)
	fmt.Printf("ready=%t (%T)\n", ready, ready)
	fmt.Printf("title=%q (%T)\n", title, title)
}
