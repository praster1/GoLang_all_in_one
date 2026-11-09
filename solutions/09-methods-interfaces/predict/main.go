package main

import "fmt"

type Counter struct{ N int }

func (c Counter) Read() int  { return c.N }
func (c *Counter) Add(n int) { c.N += n }

type Reader interface{ Read() int }

func main() {
	c := Counter{N: 2}
	var copied Reader = c
	var shared Reader = &c
	saved := c.Read
	c.Add(3)
	fmt.Println(c.Read(), copied.Read(), shared.Read())
	fmt.Println(saved())
	var missing *Counter
	var value Reader = missing
	fmt.Println(missing == nil, value == nil)
}
