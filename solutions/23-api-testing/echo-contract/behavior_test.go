package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestEchoContract(t *testing.T) {
	tests := []struct {
		name, input   string
		status        int
		message, code string
	}{
		{"normal", "Go", 200, "Go", ""},
		{"trimmed", " Go ", 200, "Go", ""},
		{"maximum", strings.Repeat("가", 8), 200, strings.Repeat("가", 8), ""},
		{"missing", "", 400, "메시지는 1~8글자입니다", "invalid_message"},
		{"blank", "  ", 400, "메시지는 1~8글자입니다", "invalid_message"},
		{"too long", strings.Repeat("가", 9), 400, "메시지는 1~8글자입니다", "invalid_message"},
	}
	handler := newHandler()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/echo?message="+url.QueryEscape(tc.input), nil)
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
