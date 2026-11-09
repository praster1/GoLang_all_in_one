package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Settings struct {
	Enabled *bool `json:"enabled,omitempty"`
}

func describe(settings Settings) string {
	if settings.Enabled == nil {
		return "미지정"
	}
	if *settings.Enabled {
		return "켜짐"
	}
	return "꺼짐"
}

func run() error {
	inputs := []string{
		`{}`,
		`{"enabled":false}`,
		`{"enabled":true}`,
		`{"enabled":null}`,
	}
	for _, input := range inputs {
		var settings Settings
		if err := json.Unmarshal([]byte(input), &settings); err != nil {
			return err
		}
		output, err := json.Marshal(settings)
		if err != nil {
			return err
		}
		fmt.Printf("%s => %s => %s\n", input, describe(settings), output)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
