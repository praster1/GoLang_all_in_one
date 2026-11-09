package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func sumCSV(source io.Reader) (int, error) {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = 2
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("헤더 읽기: %w", err)
	}
	if header[0] != "name" || header[1] != "minutes" {
		return 0, errors.New("헤더는 name,minutes여야 합니다")
	}
	total := 0
	count := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return total, nil
		}
		if err != nil {
			return 0, fmt.Errorf("레코드 읽기: %w", err)
		}
		count++
		if count > 1000 {
			return 0, errors.New("레코드는 1000개까지 허용합니다")
		}
		minutes, err := strconv.Atoi(record[1])
		if err != nil {
			return 0, fmt.Errorf("분 변환: %w", err)
		}
		if strings.TrimSpace(record[0]) == "" || minutes < 0 || minutes > 120 {
			return 0, errors.New("이름과 0~120분이 필요합니다")
		}
		total += minutes
	}
}

func run() error {
	input := "name,minutes\n\"Go, basics\",25\n테스트,35\n휴식,0\n"
	total, err := sumCSV(strings.NewReader(input))
	if err != nil {
		return err
	}
	fmt.Printf("합계: %d분\n", total)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
