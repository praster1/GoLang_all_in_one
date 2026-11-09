package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"unicode/utf8"
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

type CreateRequest struct {
	Title   *string `json:"title"`
	Minutes *int    `json:"minutes"`
}

type Lesson struct {
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

func decodeRequest(w http.ResponseWriter, r *http.Request) (CreateRequest, error) {
	var input CreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, 256)
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return input, err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return input, errors.New("expected one JSON value")
	}
	return input, nil
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /lessons/preview", func(w http.ResponseWriter, r *http.Request) {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			writeProblem(w, 415, "unsupported_media_type", "JSON 형식이 필요합니다")
			return
		}
		input, err := decodeRequest(w, r)
		if err != nil {
			var large *http.MaxBytesError
			if errors.As(err, &large) {
				writeProblem(w, 413, "body_too_large", "본문은 256바이트까지입니다")
			} else {
				writeProblem(w, 400, "invalid_json", "JSON 값을 확인하세요")
			}
			return
		}
		if input.Title == nil || input.Minutes == nil {
			writeProblem(w, 422, "missing_field", "title과 minutes가 필요합니다")
			return
		}
		title := strings.TrimSpace(*input.Title)
		if n := utf8.RuneCountInString(title); n < 1 || n > 40 {
			writeProblem(w, 422, "invalid_title", "제목은 1~40글자입니다")
			return
		}
		if *input.Minutes < 1 || *input.Minutes > 180 {
			writeProblem(w, 422, "invalid_minutes", "분량은 1~180입니다")
			return
		}
		writeJSON(w, 200, Lesson{Title: title, Minutes: *input.Minutes})
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/lessons/preview", `{"title":" Go ","minutes":25}`},
		{http.MethodPost, "/lessons/preview", `{"title":"Go"}`},
		{http.MethodPost, "/lessons/preview", `{"title":"Go","minutes":0}`},
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
