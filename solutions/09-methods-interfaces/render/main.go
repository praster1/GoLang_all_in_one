package main

import "fmt"

type Describer interface{ Describe() string }
type Note struct{ Text string }

func (note *Note) Describe() string {
	if note == nil {
		return "메모 없음"
	}
	return "메모: " + note.Text
}

type Number int

func (number Number) Describe() string { return fmt.Sprintf("숫자: %d", number) }

func render(values []Describer) ([]string, bool) {
	for _, value := range values {
		if value == nil {
			return nil, false
		}
	}
	var result []string
	for _, value := range values {
		result = append(result, value.Describe())
	}
	return result, true
}

func main() {
	var missing *Note
	lines, ok := render([]Describer{&Note{Text: "Go"}, Number(7), missing})
	fmt.Println("성공:", ok)
	for _, line := range lines {
		fmt.Println(line)
	}
	_, ok = render([]Describer{nil})
	fmt.Println("nil 인터페이스:", ok)
}
