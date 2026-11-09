package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

func parseAll(inputs []string) ([]int, error) {
	if len(inputs) > 16 {
		return nil, errors.New("입력은 16개까지 허용합니다")
	}
	results := make([]int, len(inputs))
	failures := make([]error, len(inputs))
	var group sync.WaitGroup
	for index, input := range inputs {
		group.Add(1)
		go func(index int, input string) {
			defer group.Done()
			value, err := strconv.Atoi(input)
			if err != nil {
				failures[index] = fmt.Errorf("항목 %d: %w", index, err)
				return
			}
			results[index] = value
		}(index, input)
	}
	group.Wait()
	if err := errors.Join(failures...); err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	results, err := parseAll([]string{"10", "x", "30"})
	var numberError *strconv.NumError
	fmt.Println("변환 실패:", err != nil)
	fmt.Println("실패 결과 없음:", results == nil)
	fmt.Println("숫자 오류 확인:", errors.As(err, &numberError))
}
