package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
		}{"ok"}); err != nil {
			// 응답을 이미 쓰기 시작했으므로 두 번째 오류 응답을 덧붙이지 않습니다.
			return
		}
	})
	return mux
}

func main() {
	recorder := httptest.NewRecorder()
	newHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	fmt.Println(recorder.Code, strings.TrimSpace(recorder.Body.String()))
}
