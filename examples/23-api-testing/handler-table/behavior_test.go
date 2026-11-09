package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGreetingContract(t *testing.T) {
	tests := []struct {
		name, input   string
		status        int
		message, code string
	}{
		{"normal", "Go", 200, "안녕하세요, Go!", ""},
		{"trimmed", " Go ", 200, "안녕하세요, Go!", ""},
		{"maximum", strings.Repeat("가", 20), 200, "안녕하세요, " + strings.Repeat("가", 20) + "!", ""},
		{"missing", "", 400, "이름은 1~20글자입니다", "invalid_name"},
		{"blank", "  ", 400, "이름은 1~20글자입니다", "invalid_name"},
		{"too long", strings.Repeat("가", 21), 400, "이름은 1~20글자입니다", "invalid_name"},
	}
	handler := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/greet?name="+url.QueryEscape(tc.input), nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status=%d, want %d", recorder.Code, tc.status)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type=%q", got)
			}
			var body struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid response JSON: %v", err)
			}
			if body.Message != tc.message || body.Code != tc.code {
				t.Errorf("body=%+v, want message=%q code=%q", body, tc.message, tc.code)
			}
		})
	}
}
