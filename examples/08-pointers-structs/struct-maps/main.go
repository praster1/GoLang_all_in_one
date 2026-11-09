package main

import "fmt"

type Stock struct {
	Quantity int
}

func addStock(stock map[string]Stock, name string, amount int) bool {
	if stock == nil || amount < 1 || amount > 100 {
		return false
	}
	item, ok := stock[name]
	if !ok || item.Quantity < 0 || item.Quantity > 100-amount {
		return false
	}
	item.Quantity += amount
	stock[name] = item
	return true
}

func main() {
	stock := map[string]Stock{"pen": {Quantity: 3}}
	fmt.Println("입고:", addStock(stock, "pen", 2))
	fmt.Println("맵 수량:", stock["pen"].Quantity)
	copied := stock["pen"]
	copied.Quantity++
	fmt.Println("복사본과 맵:", copied.Quantity, stock["pen"].Quantity)
	fmt.Println("없는 상품:", addStock(stock, "book", 1))
}
