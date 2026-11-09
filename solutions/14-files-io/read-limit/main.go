package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var errTooLarge = errors.New("크기 제한 초과")

func readBounded(source io.Reader, maximum int64) ([]byte, error) {
	if maximum < 0 || maximum > 1024*1024 {
		return nil, fmt.Errorf("제한은 0부터 1048576바이트까지 지정하세요")
	}
	data, err := io.ReadAll(io.LimitReader(source, maximum+1))
	if err != nil {
		return nil, fmt.Errorf("읽기: %w", err)
	}
	if int64(len(data)) > maximum {
		return nil, errTooLarge
	}
	return data, nil
}

func main() {
	data, err := readBounded(strings.NewReader("hello"), 5)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("내용: %s\n", data)
	_, err = readBounded(strings.NewReader("hello!"), 5)
	fmt.Println("초과 확인:", errors.Is(err, errTooLarge))
}
