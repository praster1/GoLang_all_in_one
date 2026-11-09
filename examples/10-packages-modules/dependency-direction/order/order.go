package order

// Basket stores a validated quantity and unit price for this exercise.
type Basket struct{ quantity, price int }

func New(quantity, price int) (Basket, bool) {
	if quantity < 0 || quantity > 1000 || price < 0 || price > 1000000 {
		return Basket{}, false
	}
	return Basket{quantity: quantity, price: price}, true
}
func (b Basket) Total() int { return b.quantity * b.price }
