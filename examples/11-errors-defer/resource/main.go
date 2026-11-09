package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type trackedReader struct {
	io.Reader
	closeErr error
	closed   bool
}

func (r *trackedReader) Close() error { r.closed = true; return r.closeErr }

// readAndClose owns input and closes it once. Any error discards the result.
func readAndClose(input io.ReadCloser) (text string, err error) {
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close input: %w", closeErr))
			text = ""
		}
	}()
	data, readErr := io.ReadAll(input)
	if readErr != nil {
		return "", fmt.Errorf("read input: %w", readErr)
	}
	return string(data), nil
}
func main() {
	input := &trackedReader{Reader: strings.NewReader("Go")}
	value, err := readAndClose(input)
	fmt.Println("결과:", value, "성공:", err == nil, "닫힘:", input.closed)
	closeFailure := errors.New("close failed")
	input = &trackedReader{Reader: strings.NewReader("Go"), closeErr: closeFailure}
	value, err = readAndClose(input)
	fmt.Println("실패 결과:", value == "", "닫기 오류:", errors.Is(err, closeFailure))
}
