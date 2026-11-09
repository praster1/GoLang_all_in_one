package main

import (
	"cmp"
	"fmt"
	"math"
)

// smallest uses cmp.Compare's total ordering, including its NaN policy.
func smallest[T cmp.Ordered](values []T) (T, bool) {
	if len(values) == 0 {
		var zero T
		return zero, false
	}
	result := values[0]
	for _, value := range values[1:] {
		if cmp.Compare(value, result) < 0 {
			result = value
		}
	}
	return result, true
}
func main() {
	n, ok := smallest([]int{8, 3, 5})
	fmt.Println("정수:", n, ok)
	word, ok := smallest([]string{"banana", "apple"})
	fmt.Println("문자열:", word, ok)
	number, _ := smallest([]float64{2, math.NaN(), 1})
	fmt.Println("NaN 정책:", math.IsNaN(number))
}
