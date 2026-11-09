package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

//go:embed defaults.json
var defaults string

type config struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func parseConfig(text string) (config, error) {
	var cfg config
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return config{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return config{}, errors.New("expected one JSON value")
	}
	if strings.TrimSpace(cfg.Name) == "" || cfg.Port < 1 || cfg.Port > 65535 {
		return config{}, errors.New("invalid config values")
	}
	return cfg, nil
}

func main() {
	cfg, err := parseConfig(defaults)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%s:%d\n", cfg.Name, cfg.Port)
}
