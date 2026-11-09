package batch

import (
	"encoding/json"
	"io"
)

type Summary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Canceled  int `json:"canceled"`
	Attempts  int `json:"attempts"`
}

type Item struct {
	Index    int    `json:"index"`
	Input    string `json:"input"`
	Output   string `json:"output,omitempty"`
	State    State  `json:"state"`
	Attempts int    `json:"attempts"`
	Error    string `json:"error,omitempty"`
}

type Report struct {
	Summary Summary `json:"summary"`
	Items   []Item  `json:"items"`
}

func NewReport(results []Result) Report {
	report := Report{Items: make([]Item, len(results))}
	report.Summary.Total = len(results)
	for i, result := range results {
		item := Item{Index: result.Job.Index, Input: result.Job.Input, Output: result.Output, State: result.State, Attempts: result.Attempts}
		if result.Err != nil {
			item.Error = result.Err.Error()
		}
		report.Items[i] = item
		report.Summary.Attempts += result.Attempts
		switch result.State {
		case Succeeded:
			report.Summary.Succeeded++
		case Failed:
			report.Summary.Failed++
		case Canceled:
			report.Summary.Canceled++
		}
	}
	return report
}

func WriteJSON(w io.Writer, results []Result) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(NewReport(results))
}
