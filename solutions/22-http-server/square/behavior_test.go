package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestSquareContract(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		path                string
		status, input, want int
	}{{"/squares/0", 200, 0, 0}, {"/squares/1", 200, 1, 1}, {"/squares/12", 200, 12, 144}, {"/squares/1000", 200, 1000, 1000000}, {"/squares/-1", 400, 0, 0}, {"/squares/1001", 400, 0, 0}, {"/squares/x", 400, 0, 0}, {"/squares/999999999999999999999999", 400, 0, 0}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.status {
			t.Fatalf("path=%s code=%d", tc.path, w.Code)
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("missing JSON header")
		}
		if tc.status == 200 {
			var result struct {
				Input  int `json:"input"`
				Square int `json:"square"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Input != tc.input || result.Square != tc.want {
				t.Errorf("path=%s result=%+v err=%v", tc.path, result, err)
			}
		} else {
			var problem Problem
			if err := json.Unmarshal(w.Body.Bytes(), &problem); err != nil || problem.Error != "invalid_number" {
				t.Errorf("problem=%+v err=%v", problem, err)
			}
		}
	}
}
