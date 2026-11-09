package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var errTag = errors.New("invalid tag")

func normalizeTags(text string) ([]string, error) {
	seen := make(map[string]struct{})
	for _, part := range strings.Split(text, ",") {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		for _, ch := range tag {
			valid := ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-'
			if !valid {
				return nil, errTag
			}
		}
		seen[strings.ToLower(tag)] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for tag := range seen {
		result = append(result, tag)
	}
	slices.Sort(result)
	return result, nil
}

func main() {
	tags, err := normalizeTags(" Go,api, GO, test-1 ")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(tags)
}
