package main

import "testing"

func TestCompleteByTitle(t *testing.T) {
	tasks := []Task{{Title: "a"}, {Title: "a"}, {Title: "b"}}
	if !completeByTitle(tasks, "a") || !tasks[0].Done || tasks[1].Done {
		t.Fatalf("first match: %+v", tasks)
	}
	if completeByTitle(tasks, "missing") || completeByTitle(nil, "a") {
		t.Fatal("missing accepted")
	}
	if !completeByTitle(tasks, "a") {
		t.Fatal("already complete should still be found")
	}
}

func Example_main() {
	main()
	// Output:
	// 복사본 수정 뒤: false false
	// 검색 결과: true
	// 읽기: false
	// 실행: true
}
