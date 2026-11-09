package report

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"
)

func ordered(summary Summary) Summary {
	summary.Groups = append([]Group{}, summary.Groups...)
	slices.SortFunc(summary.Groups, func(a, b Group) int { return strings.Compare(a.Category, b.Category) })
	return summary
}

func WriteCSV(output io.Writer, summary Summary) error {
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"category", "count", "total"}); err != nil {
		return err
	}
	for _, group := range ordered(summary).Groups {
		if err := writer.Write([]string{group.Category, strconv.Itoa(group.Count), strconv.FormatInt(group.Total, 10)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func WriteJSON(output io.Writer, summary Summary) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ordered(summary))
}
