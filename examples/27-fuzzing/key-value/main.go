package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var errRecord = errors.New("invalid record")

func parseRecord(text string) (string, string, error) {
	if len(text) > 4096 || !utf8.ValidString(text) || strings.ContainsAny(text, "\r\n") {
		return "", "", errRecord
	}
	key, value, ok := strings.Cut(text, "=")
	if !ok || key == "" {
		return "", "", errRecord
	}
	for _, ch := range key {
		if (ch < 'a' || ch > 'z') && ch != '_' {
			return "", "", errRecord
		}
	}
	return key, value, nil
}

func main() {
	key, value, err := parseRecord("title=Go=입문")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("키: %s\n값: %s\n", key, value)
}
