package main

import "fmt"

func main() {
	base := []int{10, 20, 30, 40, 50}
	middle := base[1:4]
	middle[0] = 99
	fmt.Println("원본:", base)
	fmt.Printf("부분: %v, 길이: %d, 용량: %d\n", middle, len(middle), cap(middle))
	expanded := middle[:4]
	fmt.Println("범위 확장:", expanded)
	limited := base[1:3:3]
	fmt.Printf("용량 제한: %v, 길이: %d, 용량: %d\n", limited, len(limited), cap(limited))
}
