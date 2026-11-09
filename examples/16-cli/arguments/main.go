package main

import (
	"fmt"
	"io"
	"os"
)

func showArgs(args []string, output io.Writer) error {
	if _, err := fmt.Fprintf(output, "인자 수: %d\n", len(args)); err != nil {
		return err
	}
	for index, argument := range args {
		if _, err := fmt.Fprintf(output, "%d: %q\n", index, argument); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := showArgs(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
