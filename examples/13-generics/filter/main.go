package main

import "fmt"

// filter preserves the slice type and nilness; keep must be non-nil.
func filter[S ~[]E, E any](values S, keep func(E) bool) S {
	if values == nil {
		return nil
	}
	result := make(S, 0, len(values))
	for _, value := range values {
		if keep(value) {
			result = append(result, value)
		}
	}
	return result
}

type Scores []int

func (scores Scores) Label() string { return fmt.Sprintf("점수 %d개", len(scores)) }
func main() {
	values := Scores{30, 80, 95}
	passed := filter(values, func(score int) bool { return score >= 60 })
	fmt.Println("통과:", passed, passed.Label())
	passed[0] = 100
	fmt.Println("원본:", values)
	var empty Scores
	fmt.Println("nil 유지:", filter(empty, func(int) bool { return true }) == nil)
}
