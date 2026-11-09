package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func run(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("hello", flag.ContinueOnError)
	fs.SetOutput(errOut)
	name := fs.String("name", "Go", "greeting name")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() != 0 || strings.TrimSpace(*name) == "" {
		fmt.Fprintln(errOut, "name must be non-empty; positional arguments are not accepted")
		return 2
	}
	if _, err := fmt.Fprintf(out, "안녕하세요, %s!\n", *name); err != nil {
		fmt.Fprintln(errOut, "write:", err)
		return 1
	}
	return 0
}

func main() {
	if code := run(os.Args[1:], os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}
