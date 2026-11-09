package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

var errName = errors.New("name must be 1..20 UTF-8 runes without control characters")

func writeRoster(output io.Writer, names []string) error {
	cleaned := make([]string, len(names))
	for i, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 20 {
			return fmt.Errorf("name %d: %w", i+1, errName)
		}
		for _, ch := range name {
			if unicode.IsControl(ch) {
				return fmt.Errorf("name %d: %w", i+1, errName)
			}
		}
		cleaned[i] = name
	}
	if _, err := fmt.Fprintf(output, "학습자: %d\n", len(cleaned)); err != nil {
		return fmt.Errorf("write heading: %w", err)
	}
	for _, name := range cleaned {
		if _, err := fmt.Fprintf(output, "- %s\n", name); err != nil {
			return fmt.Errorf("write name: %w", err)
		}
	}
	return nil
}

func main() {
	var output bytes.Buffer
	if err := writeRoster(&output, []string{" 민수 ", "지수"}); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(output.String())
}
