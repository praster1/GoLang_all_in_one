package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/greeting/greet"
)

func main() {
	fmt.Println(greet.Hello(" Go "))
	fmt.Println(greet.Hello(""))
}
