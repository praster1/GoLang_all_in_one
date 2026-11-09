package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func readChunks(source io.Reader) (string, error) {
	buffer := make([]byte, 4)
	var data []byte
	for {
		n, err := source.Read(buffer)
		data = append(data, buffer[:n]...)
		if errors.Is(err, io.EOF) {
			return string(data), nil
		}
		if err != nil {
			return string(data), err
		}
	}
}

func main() {
	text, err := readChunks(strings.NewReader("Go와 파일"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(text)
	fmt.Println("바이트:", len(text))
}
