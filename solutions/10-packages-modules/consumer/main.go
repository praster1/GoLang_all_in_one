package main

import (
	"fmt"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/consumer/display"
	"github.com/praster1/GoLang_all_in_one/solutions/10-packages-modules/consumer/task"
)

func main() { fmt.Println(display.Line(task.Task{Title: "패키지 복습", Done: true})) }
