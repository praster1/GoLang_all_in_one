package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/labels/label"
)

func main() {
	value, ok := label.Normalize(" GO ")
	fmt.Println("라벨:", value, ok)
	_, ok = label.Normalize("  ")
	fmt.Println("빈 값:", ok)
}
