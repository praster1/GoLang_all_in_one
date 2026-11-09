package main

import "fmt"

type Label struct{ Text string }

func (label Label) Describe() string { return label.Text }

type Task struct {
	Label
	Done bool
}

func (task Task) Describe() string {
	state := "진행"
	if task.Done {
		state = "완료"
	}
	return task.Label.Describe() + " / " + state
}
func main() {
	task := Task{Label: Label{Text: "Go 읽기"}, Done: true}
	fmt.Println(task.Describe())
	fmt.Println(task.Label.Describe())
}
