package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScoreBoundaries(t *testing.T) {
	for _, tc := range []struct {
		score, status int
		label, code   string
	}{{-1, 422, "", "invalid_score"}, {0, 200, "다시 연습", ""}, {59, 200, "다시 연습", ""}, {60, 200, "통과", ""}, {89, 200, "통과", ""}, {90, 200, "우수", ""}, {100, 200, "우수", ""}, {101, 422, "", "invalid_score"}} {
		req := httptest.NewRequest("POST", "/score", strings.NewReader(fmt.Sprintf(`{"score":%d}`, tc.score)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newHandler().ServeHTTP(w, req)
		var result struct {
			Label string `json:"label"`
			Code  string `json:"code"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		if w.Code != tc.status || err != nil || result.Label != tc.label || result.Code != tc.code {
			t.Errorf("score=%d status=%d body=%+v err=%v", tc.score, w.Code, result, err)
		}
	}
}
func TestScoreJSONContract(t *testing.T) {
	valid := `{"score":72}`
	for _, tc := range []struct {
		name, body, media string
		status            int
		code              string
	}{
		{"missing", "{}", "application/json", 422, "missing_score"}, {"null field", `{"score":null}`, "application/json", 422, "missing_score"}, {"null body", "null", "application/json", 422, "missing_score"},
		{"wrong type", `{"score":"72"}`, "application/json", 400, "invalid_json"}, {"unknown", `{"score":72,"extra":1}`, "application/json", 400, "invalid_json"},
		{"two values", valid + "{}", "application/json", 400, "invalid_json"}, {"trailing junk", valid + "x", "application/json", 400, "invalid_json"},
		{"empty", "", "application/json", 400, "invalid_json"}, {"media", valid, "text/plain", 415, "unsupported_media_type"},
		{"exact limit", valid + strings.Repeat(" ", 64-len(valid)), "application/json", 200, ""},
		{"too large", valid + strings.Repeat(" ", 65-len(valid)), "application/json", 413, "body_too_large"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/score", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.media)
			w := httptest.NewRecorder()
			newHandler().ServeHTTP(w, req)
			if w.Code != tc.status || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("status=%d header=%s body=%s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
			}
			if tc.code != "" {
				var p Problem
				if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil || p.Code != tc.code {
					t.Fatalf("problem=%+v err=%v", p, err)
				}
			}
		})
	}
}
