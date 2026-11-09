package main

import "fmt"

func unique[S ~[]E, E comparable](values S) S {
	if values == nil {
		return nil
	}
	result := make(S, 0, len(values))
	seen := make(map[E]bool)
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

type Names []string

func main() {
	fmt.Println(unique([]int{2, 1, 2, 3, 1}))
	names := unique(Names{"Go", "SQL", "Go"})
	fmt.Printf("%v (%T)\n", names, names)
}
