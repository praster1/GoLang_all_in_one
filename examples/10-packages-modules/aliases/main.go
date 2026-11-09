package main

import (
	"fmt"
	brief "github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/aliases/brief"
	detail "github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/aliases/detail"
)

func main() {
	fmt.Println(brief.Title("메모"))
	fmt.Println(detail.Title("메모", 3))
}
