package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"
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

func inspect(client *http.Client, address, method string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1025))
	if err != nil {
		return 0, "", err
	}
	if len(body) > 1024 {
		return 0, "", errors.New("response too large")
	}
	return resp.StatusCode, string(body), nil
}

func run() error {
	handler := newHandler(memoryStore{lessons: map[int]Lesson{1: {ID: 1, Title: "Go"}}})
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path string }{{"GET", "/lessons/1"}, {"HEAD", "/lessons/1"}, {"GET", "/lessons/2"}} {
		code, body, err := inspect(client, server.URL+tc.path, tc.method)
		if err != nil {
			return err
		}
		if body == "" {
			fmt.Printf("%s %d (빈 본문)\n", tc.method, code)
		} else {
			fmt.Printf("%s %d %s\n", tc.method, code, strings.TrimSpace(body))
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
