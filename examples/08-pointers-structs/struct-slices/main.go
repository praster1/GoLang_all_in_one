package main

import "fmt"

type Task struct {
	Title string
	Done  bool
}

func completeByTitle(tasks []Task, title string) bool {
	for i := range tasks {
		if tasks[i].Title == title {
			tasks[i].Done = true
			return true
		}
	}
	return false
}

func main() {
	tasks := []Task{{Title: "읽기"}, {Title: "실행"}}
	for _, task := range tasks {
		task.Done = true
	}
	fmt.Println("복사본 수정 뒤:", tasks[0].Done, tasks[1].Done)
	fmt.Println("검색 결과:", completeByTitle(tasks, "실행"))
	for _, task := range tasks {
		fmt.Printf("%s: %t\n", task.Title, task.Done)
	}
}
