package main

import "fmt"

type Score struct{ Points int }

func (score Score) Added(amount int) Score {
	score.Points += amount
	return score
}

func (score Score) Label() string {
	return fmt.Sprintf("%d점", score.Points)
}

func main() {
	original := Score{Points: 70}
	changed := original.Added(5)
	fmt.Println("원본:", original.Label())
	fmt.Println("새 값:", changed.Label())
}
