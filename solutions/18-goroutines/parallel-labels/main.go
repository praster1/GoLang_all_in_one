package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

func labels(names []string) ([]string, error) {
	if len(names) > 16 {
		return nil, errors.New("이름은 16개까지 허용합니다")
	}
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			return nil, errors.New("빈 이름은 허용하지 않습니다")
		}
	}
	results := make([]string, len(names))
	var group sync.WaitGroup
	for index, name := range names {
		group.Go(func() {
			results[index] = strings.ToUpper(strings.TrimSpace(name))
		})
	}
	group.Wait()
	return results, nil
}

func main() {
	results, err := labels([]string{" go ", "test", "한글"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(results)
}
