package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

func run() error {
	var destination bytes.Buffer
	writer := bufio.NewWriter(&destination)
	if _, err := fmt.Fprintln(writer, "Go"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "buffer"); err != nil {
		return err
	}
	fmt.Println("Flush 전:", destination.Len())
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("버퍼 비우기: %w", err)
	}
	_, err := fmt.Printf("Flush 후: %d\n%s", destination.Len(), destination.String())
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
