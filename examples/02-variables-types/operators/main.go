package main

import "fmt"

func main() {
	a, b := 7, 2

	fmt.Println("합:", a+b)
	fmt.Println("차:", a-b)
	fmt.Println("곱:", a*b)
	fmt.Println("몫:", a/b)
	fmt.Println("나머지:", a%b)
	fmt.Println("비교:", a > b, a == b)
	fmt.Println("조건:", a > 0 && b > 0, a == 0 || b == 0, !(a == b))
	fmt.Println("우선순위:", 2+3*4, (2+3)*4)
}
