package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Student struct {
	Name  string
	Score int
}

func newStudent(name string, score int) (Student, bool) {
	if !utf8.ValidString(name) || score < 0 || score > 100 {
		return Student{}, false
	}
	cleaned := strings.TrimSpace(name)
	if cleaned == "" {
		return Student{}, false
	}
	return Student{Name: cleaned, Score: score}, true
}

func main() {
	student, ok := newStudent(" Mina ", 80)
	fmt.Printf("생성: %t, 이름: %q, 점수: %d\n", ok, student.Name, student.Score)
	invalid, ok := newStudent("Jin", -1)
	fmt.Printf("잘못된 점수: %t, 값: %+v\n", ok, invalid)
}
