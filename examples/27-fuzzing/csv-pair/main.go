package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
)

var errCarriageReturn = errors.New("carriage return not supported")

func encodePair(left, right string) (string, error) {
	if strings.ContainsAny(left+right, "\r") {
		return "", errCarriageReturn
	}
	var result strings.Builder
	w := csv.NewWriter(&result)
	if err := w.Write([]string{left, right}); err != nil {
		return "", err
	}
	w.Flush()
	return result.String(), w.Error()
}

func main() {
	text, err := encodePair("Go,입문", "한글")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(text)
}
