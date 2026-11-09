package main

import "testing"

func TestNewStudent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		score int
		want  Student
		ok    bool
	}{
		{" Mina ", 80, Student{Name: "Mina", Score: 80}, true}, {"\u3000한글\t", 0, Student{Name: "한글"}, true},
		{"Go", 100, Student{Name: "Go", Score: 100}, true}, {"", 50, Student{}, false}, {" \t\n", 50, Student{}, false},
		{"Go", -1, Student{}, false}, {"Go", 101, Student{}, false}, {"\xff", 50, Student{}, false},
	} {
		got, ok := newStudent(tc.name, tc.score)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("case %+v: %+v %t", tc, got, ok)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 생성: true, 이름: "Mina", 점수: 80
	// 잘못된 점수: false, 값: {Name: Score:0}
}
