package main

import (
	"errors"
	"fmt"
)

func parseClock(text string) (int, error) {
	if len(text) != 5 || text[2] != ':' {
		return 0, errors.New("expected HH:MM")
	}
	for _, i := range []int{0, 1, 3, 4} {
		if text[i] < '0' || text[i] > '9' {
			return 0, errors.New("expected ASCII digits")
		}
	}
	hours := int(text[0]-'0')*10 + int(text[1]-'0')
	minutes := int(text[3]-'0')*10 + int(text[4]-'0')
	if hours > 23 || minutes > 59 {
		return 0, errors.New("time out of range")
	}
	return hours*60 + minutes, nil
}

func main() {
	minutes, err := parseClock("09:30")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("경과 분:", minutes)
}
