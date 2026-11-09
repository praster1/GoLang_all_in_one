package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/dependency-direction/order"
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/dependency-direction/report"
)

func main() {
	basket, ok := order.New(3, 2000)
	if !ok {
		fmt.Println("입력이 잘못되었습니다")
		return
	}
	fmt.Println(report.Summary(basket))
}
