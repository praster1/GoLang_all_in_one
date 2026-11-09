package main

import "fmt"

func partition[S ~[]E, E any](values S, keep func(E) bool) (S, S) {
	if values == nil {
		return nil, nil
	}
	yes, no := make(S, 0, len(values)), make(S, 0, len(values))
	for _, value := range values {
		if keep(value) {
			yes = append(yes, value)
		} else {
			no = append(no, value)
		}
	}
	return yes, no
}
func main() {
	even, odd := partition([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 })
	fmt.Println("짝수:", even)
	fmt.Println("홀수:", odd)
}
