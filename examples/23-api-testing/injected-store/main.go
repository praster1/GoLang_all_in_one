package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

type Problem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}

func writeProblem(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Problem{Code: code, Message: message})
}

var ErrNotFound = errors.New("lesson not found")

type Lesson struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type LessonStore interface {
	Lookup(context.Context, int) (Lesson, error)
}

type memoryStore struct{ lessons map[int]Lesson }

func (s memoryStore) Lookup(ctx context.Context, id int) (Lesson, error) {
	if err := ctx.Err(); err != nil {
		return Lesson{}, err
	}
	lesson, ok := s.lessons[id]
	if !ok {
		return Lesson{}, ErrNotFound
	}
	return lesson, nil
}

func newHandler(store LessonStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lessons/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 {
			writeProblem(w, 400, "invalid_id", "ID는 양수 정수입니다")
			return
		}
		lesson, err := store.Lookup(r.Context(), id)
		if err != nil {
			if r.Context().Err() != nil {
				return
			}
			if errors.Is(err, ErrNotFound) {
				writeProblem(w, 404, "not_found", "학습을 찾을 수 없습니다")
			} else {
				writeProblem(w, 500, "internal_error", "요청을 처리하지 못했습니다")
			}
			return
		}
		writeJSON(w, 200, lesson)
	})
	return mux
}

func main() {
	handler := newHandler(memoryStore{lessons: map[int]Lesson{1: {ID: 1, Title: "Go"}}})
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodGet, "/lessons/1", ""},
		{http.MethodGet, "/lessons/2", ""},
		{http.MethodGet, "/lessons/x", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		fmt.Printf("%d %s\n", recorder.Code, strings.TrimSpace(recorder.Body.String()))
	}
}
