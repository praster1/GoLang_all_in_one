package main

import "fmt"

type Counts struct{ Ints, Strings, Bools, Others int }

func countTypes(values []any) Counts {
	var result Counts
	for _, value := range values {
		switch value.(type) {
		case int:
			result.Ints++
		case string:
			result.Strings++
		case bool:
			result.Bools++
		default:
			result.Others++
		}
	}
	return result
}

func main() {
	values := []any{1, "Go", true, int64(2), nil, "한글", false}
	result := countTypes(values)
	fmt.Printf("int: %d, string: %d, bool: %d, 기타: %d\n", result.Ints, result.Strings, result.Bools, result.Others)
}
