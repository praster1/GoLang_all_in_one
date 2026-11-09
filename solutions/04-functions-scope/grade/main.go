package main

import "fmt"

func classifyScore(score int) (string, bool) {
	if score < 0 || score > 100 {
		return "", false
	}
	if score >= 90 {
		return "우수", true
	}
	if score >= 60 {
		return "통과", true
	}
	return "다시 연습", true
}

func main() {
	score := 72
	label, ok := classifyScore(score)
	if !ok {
		fmt.Println("점수 확인")
		return
	}
	fmt.Printf("%d점: %s\n", score, label)
}
