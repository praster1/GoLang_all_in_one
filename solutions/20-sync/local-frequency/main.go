package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

func countWords(parts [][]string) (map[string]int, error) {
	if len(parts) > 8 {
		return nil, errors.New("묶음은 8개까지 허용합니다")
	}
	total := 0
	for _, part := range parts {
		if len(part) > 1000-total {
			return nil, errors.New("전체 단어는 1000개까지 허용합니다")
		}
		total += len(part)
		for _, word := range part {
			if len(word) > 64 || !utf8.ValidString(word) || strings.TrimSpace(word) == "" {
				return nil, errors.New("단어는 공백이 아닌 유효한 UTF-8, 64바이트 이하여야 합니다")
			}
		}
	}
	partials := make([]map[string]int, len(parts))
	var group sync.WaitGroup
	for index, part := range parts {
		group.Go(func() {
			local := map[string]int{}
			for _, word := range part {
				local[strings.ToLower(strings.TrimSpace(word))]++
			}
			partials[index] = local
		})
	}
	group.Wait()
	result := map[string]int{}
	for _, local := range partials {
		for word, count := range local {
			result[word] += count
		}
	}
	return result, nil
}
func main() {
	counts, err := countWords([][]string{{"Go", "test"}, {"go", "한글"}, {" test ", "GO"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, counts[key])
	}
}
