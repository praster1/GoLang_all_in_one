package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"slices"
	"strings"
)

func Analyze(input io.Reader, options Options) (Summary, error) {
	if err := options.Validate(); err != nil {
		return Summary{}, err
	}
	limited := &io.LimitedReader{R: input, N: options.MaxBytes + 1}
	reader := csv.NewReader(limited)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if limited.N == 0 {
		return Summary{}, fmt.Errorf("%w: max-bytes=%d", ErrLimit, options.MaxBytes)
	}
	if err != nil {
		if err == io.EOF {
			return Summary{}, fmt.Errorf("%w: missing header", ErrInput)
		}
		return Summary{}, fmt.Errorf("read header: %w", err)
	}
	if !slices.Equal(header, []string{"category", "amount"}) {
		return Summary{}, fmt.Errorf("%w: header must be category,amount", ErrInput)
	}
	groups := make(map[string]Group)
	summary := Summary{Groups: make([]Group, 0)}
	for recordNumber := 2; ; recordNumber++ {
		fields, err := reader.Read()
		if limited.N == 0 {
			return Summary{}, fmt.Errorf("%w: max-bytes=%d", ErrLimit, options.MaxBytes)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return Summary{}, fmt.Errorf("read record %d: %w", recordNumber, err)
		}
		if summary.Rows == options.MaxRows {
			return Summary{}, fmt.Errorf("%w: max-rows=%d", ErrLimit, options.MaxRows)
		}
		record, err := ParseRow(fields, recordNumber)
		if err != nil {
			return Summary{}, err
		}
		group := groups[record.Category]
		group.Category = record.Category
		group.Count++
		group.Total += record.Amount
		groups[record.Category] = group
		summary.Rows++
		summary.Total += record.Amount
	}
	for _, group := range groups {
		summary.Groups = append(summary.Groups, group)
	}
	slices.SortFunc(summary.Groups, func(a, b Group) int { return strings.Compare(a.Category, b.Category) })
	return summary, nil
}
