package main

import (
	"errors"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name, title string
		minutes     int
		field       string
	}{
		{"normal", " Go ", 25, ""}, {"one rune", "가", 1, ""}, {"forty runes", strings.Repeat("가", 40), 180, ""},
		{"blank", "  ", 25, "title"}, {"too long", strings.Repeat("가", 41), 25, "title"}, {"zero minutes", "Go", 0, "minutes"}, {"excess minutes", "Go", 181, "minutes"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := CreateInput{Title: tc.title, Minutes: tc.minutes}
			got, err := validate(original)
			if tc.field == "" {
				if err != nil || got.Title != strings.TrimSpace(tc.title) || got.Minutes != tc.minutes {
					t.Fatalf("got=%+v err=%v", got, err)
				}
			} else {
				var field *FieldError
				if !errors.As(err, &field) || field.Field != tc.field {
					t.Fatalf("error=%v want field=%s", err, tc.field)
				}
			}
			if original.Title != tc.title {
				t.Error("caller input changed")
			}
		})
	}
}
