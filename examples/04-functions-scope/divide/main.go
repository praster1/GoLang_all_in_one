package main

import "fmt"

func divide(dividend, divisor int) (int, int, bool) {
	if dividend < 0 || divisor <= 0 {
		return 0, 0, false
	}
	return dividend / divisor, dividend % divisor, true
}

func main() {
	quotient, remainder, ok := divide(17, 5)
	if !ok {
		fmt.Println("입력 확인")
		return
	}
	fmt.Printf("몫: %d, 나머지: %d\n", quotient, remainder)

	_, _, ok = divide(17, 0)
	fmt.Printf("0으로 나누기 가능: %t\n", ok)
}
