package main

import "fmt"

func main() {
	total := 7
	count := 2
	var offset int64 = 3

	fmt.Printf("형을 맞춘 합: %d\n", int64(total)+offset)
	fmt.Printf("정수 나눗셈: %d\n", total/count)
	fmt.Printf("나눈 뒤 변환: %.1f\n", float64(total/count))
	fmt.Printf("나누기 전 변환: %.1f\n", float64(total)/float64(count))

	distance := 3.9
	fmt.Printf("소수 부분 제거: %d\n", int(distance))
}
