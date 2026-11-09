package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

func newHandler() http.Handler {
	var visits atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("POST /visits", func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			Count int64 `json:"count"`
		}{Count: visits.Add(1)}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("GET /visits", func(w http.ResponseWriter, r *http.Request) {
		result := struct {
			Count int64 `json:"count"`
		}{Count: visits.Load()}
		writeJSON(w, 200, result)
	})
	return mux
}

func main() {
	handler := newHandler()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/visits", ""},
		{http.MethodPost, "/visits", ""},
		{http.MethodPost, "/visits", ""},
		{http.MethodGet, "/visits", ""},
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
