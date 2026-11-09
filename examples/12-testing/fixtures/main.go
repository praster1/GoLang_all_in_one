package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

func readLabel(input io.Reader) (string, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return "", fmt.Errorf("read label: %w", err)
	}
	label := strings.TrimSpace(string(data))
	if !utf8.ValidString(label) || label == "" || utf8.RuneCountInString(label) > 20 || strings.ContainsAny(label, "\r\n") {
		return "", errors.New("label must be one line of 1..20 UTF-8 runes")
	}
	return label, nil
}

func main() {
	label, err := readLabel(strings.NewReader("  한글\n"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("라벨:", label)
}
