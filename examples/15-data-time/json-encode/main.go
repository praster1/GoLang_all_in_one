package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Study struct {
	Name    string `json:"name"`
	Minutes int    `json:"minutes"`
	Done    bool   `json:"done"`
}

func run() error {
	study := Study{Name: "Go", Minutes: 25, Done: false}
	data, err := json.Marshal(study)
	if err != nil {
		return fmt.Errorf("JSON 인코딩: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
