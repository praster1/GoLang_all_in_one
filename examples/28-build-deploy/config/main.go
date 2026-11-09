package main

import (
	"errors"
	"fmt"
	"strconv"
)

type config struct{ Port int }

func loadConfig(lookup func(string) (string, bool)) (config, error) {
	value, ok := lookup("BOOK_PORT")
	if !ok {
		return config{Port: 8080}, nil
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return config{}, errors.New("BOOK_PORT must be 1..65535")
	}
	return config{Port: port}, nil
}

func main() {
	values := map[string]string{"BOOK_PORT": "9090"}
	cfg, err := loadConfig(func(key string) (string, bool) { value, ok := values[key]; return value, ok })
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("실행 설정:", cfg.Port)
}
