// Package todo defines the task domain and interchangeable repositories.
package todo

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrNotFound = errors.New("할 일을 찾을 수 없습니다")
var ErrInvalidID = errors.New("ID는 양수여야 합니다")
var ErrInvalidPage = errors.New("after는 0 이상, limit는 1~100이어야 합니다")
var ErrIDExhausted = errors.New("새 ID를 발급할 수 없습니다")

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

func NormalizeTitle(title string) (string, error) {
	if !utf8.ValidString(title) {
		return "", &ValidationError{"title", "올바른 UTF-8이어야 합니다"}
	}
	title = strings.TrimSpace(title)
	if n := utf8.RuneCountInString(title); n < 1 || n > 80 {
		return "", &ValidationError{"title", "공백 제거 후 1~80룬이어야 합니다"}
	}
	for _, r := range title {
		if unicode.IsControl(r) {
			return "", &ValidationError{"title", "제어 문자를 포함할 수 없습니다"}
		}
	}
	return title, nil
}
func validatePage(after int64, limit int) error {
	if after < 0 || limit < 1 || limit > 100 {
		return ErrInvalidPage
	}
	return nil
}

// Repository returns tasks by value. List is sorted by ID, ascending.
// Create never reuses a previously issued ID during this repository's lifetime.
// Canceled contexts must be observed, but cancellation does not undo a completed write.
type Repository interface {
	Create(context.Context, string) (Task, error)
	Get(context.Context, int64) (Task, error)
	List(context.Context, int64, int) ([]Task, error)
	Replace(context.Context, int64, string, bool) (Task, error)
	Delete(context.Context, int64) error
}
