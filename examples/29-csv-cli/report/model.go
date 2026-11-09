package report

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInput = errors.New("invalid CSV input")
var ErrLimit = errors.New("input limit exceeded")

type DataError struct {
	Record int
	Field  string
	Err    error
}

func (e *DataError) Error() string {
	return fmt.Sprintf("record %d field %s: %v", e.Record, e.Field, e.Err)
}
func (e *DataError) Unwrap() error { return e.Err }

type Record struct {
	Category string
	Amount   int64
}
type Group struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Total    int64  `json:"total"`
}
type Summary struct {
	Rows   int     `json:"rows"`
	Total  int64   `json:"total"`
	Groups []Group `json:"groups"`
}

type Options struct {
	MaxBytes int64
	MaxRows  int
}

func DefaultOptions() Options { return Options{MaxBytes: 1 << 20, MaxRows: 10000} }
func (o Options) Validate() error {
	if o.MaxBytes < 1 || o.MaxBytes > 64<<20 {
		return errors.New("max-bytes must be 1..67108864")
	}
	if o.MaxRows < 1 || o.MaxRows > 1000000 {
		return errors.New("max-rows must be 1..1000000")
	}
	return nil
}

func ParseRow(fields []string, recordNumber int) (Record, error) {
	fail := func(field, message string) (Record, error) {
		return Record{}, &DataError{recordNumber, field, fmt.Errorf("%w: %s", ErrInput, message)}
	}
	if len(fields) != 2 {
		return fail("row", "expected 2 fields")
	}
	category := strings.TrimSpace(fields[0])
	if category == "" || !utf8.ValidString(category) || utf8.RuneCountInString(category) > 40 {
		return fail("category", "expected 1..40 valid UTF-8 runes")
	}
	for _, ch := range category {
		if unicode.IsControl(ch) {
			return fail("category", "control character is not allowed")
		}
	}
	amountText := strings.TrimSpace(fields[1])
	if amountText == "" {
		return fail("amount", "expected decimal digits")
	}
	for _, ch := range amountText {
		if ch < '0' || ch > '9' {
			return fail("amount", "expected decimal digits")
		}
	}
	amount, err := strconv.ParseInt(amountText, 10, 64)
	if err != nil || amount > 1000000000 {
		return fail("amount", "expected 0..1000000000")
	}
	return Record{category, amount}, nil
}
