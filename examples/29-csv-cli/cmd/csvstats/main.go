package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/29-csv-cli/report"
)

func run(args []string, input io.Reader, output, errOutput io.Writer) (code int) {
	defaults := report.DefaultOptions()
	fs := flag.NewFlagSet("csvstats", flag.ContinueOnError)
	fs.SetOutput(errOutput)
	path := fs.String("input", "-", "CSV path or - for standard input")
	format := fs.String("format", "csv", "csv or json")
	maxBytes := fs.Int64("max-bytes", defaults.MaxBytes, "maximum input bytes")
	maxRows := fs.Int("max-rows", defaults.MaxRows, "maximum data records")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	options := report.Options{MaxBytes: *maxBytes, MaxRows: *maxRows}
	if fs.NArg() != 0 || (*format != "csv" && *format != "json") || *path == "" {
		fmt.Fprintln(errOutput, "invalid arguments: use -input, -format csv|json, and no positional arguments")
		return 2
	}
	if err := options.Validate(); err != nil {
		fmt.Fprintln(errOutput, err)
		return 2
	}
	if *path != "-" {
		file, err := os.Open(*path)
		if err != nil {
			fmt.Fprintln(errOutput, "open:", err)
			return 1
		}
		defer func() {
			if err := file.Close(); err != nil && code == 0 {
				fmt.Fprintln(errOutput, "close:", err)
				code = 1
			}
		}()
		input = file
	}
	summary, err := report.Analyze(input, options)
	if err != nil {
		fmt.Fprintln(errOutput, "analyze:", err)
		return 1
	}
	if *format == "json" {
		err = report.WriteJSON(output, summary)
	} else {
		err = report.WriteCSV(output, summary)
	}
	if err != nil {
		fmt.Fprintln(errOutput, "write:", err)
		return 1
	}
	return 0
}

func main() {
	if code := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}
