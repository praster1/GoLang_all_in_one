package main

import "fmt"

func withdraw(stock map[string]int, item string, quantity int) bool {
	available, ok := stock[item]
	if !ok || quantity <= 0 || quantity > available {
		return false
	}
	stock[item] = available - quantity
	return true
}
func main() {
	stock := map[string]int{"pen": 5, "book": 0}
	fmt.Println("첫 출고:", withdraw(stock, "pen", 3))
	fmt.Println("남은 pen:", stock["pen"])
	fmt.Println("초과 출고:", withdraw(stock, "pen", 3))
	fmt.Println("없는 상품:", withdraw(stock, "note", 1))
	fmt.Println("최종 pen:", stock["pen"])
}
