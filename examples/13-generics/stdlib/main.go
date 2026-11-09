package main

import (
	"cmp"
	"fmt"
	"slices"
)

type Student struct {
	Name  string
	Score int
}

func main() {
	original := []Student{{Name: "민수", Score: 80}, {Name: "영희", Score: 95}, {Name: "재성", Score: 80}}
	ranked := slices.Clone(original)
	slices.SortFunc(ranked, func(a, b Student) int {
		if order := cmp.Compare(b.Score, a.Score); order != 0 {
			return order
		}
		return cmp.Compare(a.Name, b.Name)
	})
	for _, student := range ranked {
		fmt.Println(student.Name, student.Score)
	}
	fmt.Println("원본 첫째:", original[0].Name)
	fmt.Println("포함:", slices.Contains([]string{"Go", "SQL"}, "Go"))
}
