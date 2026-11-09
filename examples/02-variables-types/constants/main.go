package main

import "fmt"

func main() {
	const minutesPerHour = 60
	const course string = "Go"
	studyMinutes := 135
	hours := studyMinutes / minutesPerHour
	minutes := studyMinutes % minutesPerHour

	fmt.Printf("과정: %s\n", course)
	fmt.Printf("학습 시간: %d시간 %d분\n", hours, minutes)
}
