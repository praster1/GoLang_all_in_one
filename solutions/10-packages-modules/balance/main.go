package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/balance/wallet"
)

func main() {
	w, ok := wallet.New(3000)
	if !ok {
		fmt.Println("입력 오류")
		return
	}
	fmt.Println("사용:", w.Spend(1000), w.Balance())
	fmt.Println("초과:", w.Spend(3000), w.Balance())
}
