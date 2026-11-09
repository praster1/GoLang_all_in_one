package main

import "fmt"

type Counter struct{ N int }

func (c *Counter) Add(n int) { c.N += n }

type Adder interface{ Add(int) }

func addTwo(target Adder) { target.Add(2) }
func main() {
	c := Counter{N: 3}
	var target Adder = &c
	addTwo(target)
	fmt.Println(c.N)
}
