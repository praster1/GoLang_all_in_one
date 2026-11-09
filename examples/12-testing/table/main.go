package main

import (
	"errors"
	"fmt"
)

func shippingFee(grams int) (int, error) {
	if grams < 1 || grams > 10000 {
		return 0, errors.New("grams must be 1..10000")
	}
	switch {
	case grams <= 1000:
		return 3000, nil
	case grams <= 5000:
		return 5000, nil
	default:
		return 8000, nil
	}
}

func main() {
	fee, err := shippingFee(1200)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("배송비:", fee)
}
