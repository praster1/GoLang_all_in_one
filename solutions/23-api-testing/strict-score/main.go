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

type ScoreRequest struct {
	Score *int `json:"score"`
}

func decodeRequest(w http.ResponseWriter, r *http.Request) (ScoreRequest, error) {
	var input ScoreRequest
	r.Body = http.MaxBytesReader(w, r.Body, 64)
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

func classifyScore(score int) string {
	if score >= 90 {
		return "우수"
	}
	if score >= 60 {
		return "통과"
	}
	return "다시 연습"
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /score", func(w http.ResponseWriter, r *http.Request) {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			writeProblem(w, 415, "unsupported_media_type", "JSON 형식이 필요합니다")
			return
		}
		input, err := decodeRequest(w, r)
		if err != nil {
			var large *http.MaxBytesError
			if errors.As(err, &large) {
				writeProblem(w, 413, "body_too_large", "본문은 64바이트까지입니다")
			} else {
				writeProblem(w, 400, "invalid_json", "JSON 값을 확인하세요")
			}
			return
		}
		if input.Score == nil {
			writeProblem(w, 422, "missing_score", "score가 필요합니다")
			return
		}
		if *input.Score < 0 || *input.Score > 100 {
			writeProblem(w, 422, "invalid_score", "점수는 0~100입니다")
			return
		}
		result := struct {
			Label string `json:"label"`
		}{Label: classifyScore(*input.Score)}
		writeJSON(w, 200, result)
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/score", `{"score":0}`},
		{http.MethodPost, "/score", `{"score":72}`},
		{http.MethodPost, "/score", `{}`},
		{http.MethodPost, "/score", `{"score":null}`},
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
