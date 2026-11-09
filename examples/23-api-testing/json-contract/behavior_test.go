package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONContract(t *testing.T) {
	valid := `{"title":"Go","minutes":25}`
	for _, tc := range []struct {
		name, body, media string
		status            int
		code              string
	}{
		{"normal", valid, "application/json", 200, ""}, {"parameters", valid, "application/json; charset=utf-8", 200, ""},
		{"missing minutes", `{"title":"Go"}`, "application/json", 422, "missing_field"}, {"null minutes", `{"title":"Go","minutes":null}`, "application/json", 422, "missing_field"},
		{"null object", "null", "application/json", 422, "missing_field"}, {"zero minutes", `{"title":"Go","minutes":0}`, "application/json", 422, "invalid_minutes"},
		{"minimum", `{"title":"Go","minutes":1}`, "application/json", 200, ""}, {"maximum", `{"title":"Go","minutes":180}`, "application/json", 200, ""},
		{"excess", `{"title":"Go","minutes":181}`, "application/json", 422, "invalid_minutes"}, {"blank title", `{"title":" ","minutes":25}`, "application/json", 422, "invalid_title"},
		{"unknown", `{"title":"Go","minutes":25,"extra":1}`, "application/json", 400, "invalid_json"},
		{"two values", valid + "{}", "application/json", 400, "invalid_json"}, {"trailing junk", valid + "x", "application/json", 400, "invalid_json"},
		{"empty", "", "application/json", 400, "invalid_json"}, {"wrong type", `{"title":"Go","minutes":"25"}`, "application/json", 400, "invalid_json"},
		{"exact limit", valid + strings.Repeat(" ", 256-len(valid)), "application/json", 200, ""},
		{"too large", valid + strings.Repeat(" ", 257-len(valid)), "application/json", 413, "body_too_large"},
		{"wrong media", valid, "text/plain", 415, "unsupported_media_type"},
		{"v1 case matching", `{"TITLE":"Go","MINUTES":25}`, "application/json", 200, ""},
		{"v1 duplicate key", `{"title":"first","title":"Go","minutes":25}`, "application/json", 200, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/lessons/preview", strings.NewReader(tc.body))
			request.Header.Set("Content-Type", tc.media)
			recorder := httptest.NewRecorder()
			newHandler().ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
			}
			if recorder.Header().Get("Content-Type") != "application/json" {
				t.Fatal("wrong response media type")
			}
			if tc.status == 200 {
				var lesson Lesson
				if err := json.Unmarshal(recorder.Body.Bytes(), &lesson); err != nil || lesson.Title != "Go" {
					t.Fatalf("lesson=%+v err=%v", lesson, err)
				}
			} else {
				var problem Problem
				if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil || problem.Code != tc.code {
					t.Fatalf("problem=%+v err=%v", problem, err)
				}
			}
		})
	}
}
