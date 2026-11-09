package main

import "fmt"

type Item struct {
	Name     string
	Quantity int
}

type Order struct {
	ID    string
	Items []Item
}

func cloneOrder(order Order) Order {
	result := order
	if order.Items != nil {
		result.Items = make([]Item, len(order.Items))
		copy(result.Items, order.Items)
	}
	return result
}

func main() {
	original := Order{ID: "A-1", Items: []Item{{Name: "book", Quantity: 2}, {Name: "pen", Quantity: 3}}}
	cloned := cloneOrder(original)
	cloned.ID = "A-2"
	cloned.Items[0].Quantity = 9
	fmt.Printf("원본: %s, %s %d개\n", original.ID, original.Items[0].Name, original.Items[0].Quantity)
	fmt.Printf("복사본: %s, %s %d개\n", cloned.ID, cloned.Items[0].Name, cloned.Items[0].Quantity)
}
