package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/private-state/counter"
)

func main() {
	var c counter.Counter
	fmt.Println("추가:", c.Add(30))
	fmt.Println("초과:", c.Add(80))
	fmt.Println("현재:", c.Value())
}
