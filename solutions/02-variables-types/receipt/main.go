package main

import "fmt"

func main() {
	const unitPrice = 1200
	const shippingFee = 2500
	quantity := 3
	discount := 500
	budget := 6000

	subtotal := unitPrice * quantity
	total := subtotal - discount + shippingFee
	withinBudget := total <= budget

	fmt.Printf("상품 금액: %d원\n", subtotal)
	fmt.Printf("할인: %d원\n", discount)
	fmt.Printf("배송비: %d원\n", shippingFee)
	fmt.Printf("합계: %d원\n", total)
	fmt.Printf("예산 이내: %t\n", withinBudget)
}
