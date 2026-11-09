package main

import "fmt"

type Profile struct {
	Name   string
	Scores []int
	Labels map[string]string
}

func cloneProfile(profile Profile) Profile {
	result := profile
	if profile.Scores != nil {
		result.Scores = make([]int, len(profile.Scores))
		copy(result.Scores, profile.Scores)
	}
	if profile.Labels != nil {
		result.Labels = make(map[string]string, len(profile.Labels))
		for key, value := range profile.Labels {
			result.Labels[key] = value
		}
	}
	return result
}

func main() {
	original := Profile{Name: "Mina", Scores: []int{70, 80}, Labels: map[string]string{"team": "go"}}
	shallow := original
	shallow.Name = "Jin"
	shallow.Scores[0] = 99
	shallow.Labels["team"] = "web"
	fmt.Printf("얕은 복사 뒤 원본: %s, %v, %s\n", original.Name, original.Scores, original.Labels["team"])
	cloned := cloneProfile(original)
	cloned.Name = "Sol"
	cloned.Scores[0] = 7
	cloned.Labels["team"] = "cli"
	fmt.Printf("독립 복사 뒤 원본: %s, %v, %s\n", original.Name, original.Scores, original.Labels["team"])
	fmt.Printf("독립 복사본: %s, %v, %s\n", cloned.Name, cloned.Scores, cloned.Labels["team"])
}
