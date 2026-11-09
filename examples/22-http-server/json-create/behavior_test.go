package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateLesson(t *testing.T) {
	valid := `{"title":" Go ","minutes":25}`
	for _, tc := range []struct {
		name, body, media string
		status            int
	}{
		{"valid", valid, "application/json", 201}, {"parameters", valid, "application/json; charset=utf-8", 201},
		{"exact limit", valid + strings.Repeat(" ", 256-len(valid)), "application/json", 201},
		{"too large", valid + strings.Repeat(" ", 257-len(valid)), "application/json", 413},
		{"wrong media", valid, "text/plain", 415}, {"missing media", valid, "", 415},
		{"empty", "", "application/json", 400}, {"null", "null", "application/json", 422},
		{"unknown", `{"title":"Go","minutes":25,"x":1}`, "application/json", 400},
		{"two values", valid + "{}", "application/json", 400}, {"trailing junk", valid + "x", "application/json", 400},
		{"wrong type", `{"title":"Go","minutes":"25"}`, "application/json", 400},
		{"blank title", `{"title":" ","minutes":25}`, "application/json", 422},
		{"below", `{"title":"Go","minutes":0}`, "application/json", 422},
		{"minimum", `{"title":"Go","minutes":1}`, "application/json", 201},
		{"maximum", `{"title":"Go","minutes":180}`, "application/json", 201},
		{"above", `{"title":"Go","minutes":181}`, "application/json", 422},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/lessons", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.media)
			w := httptest.NewRecorder()
			newHandler().ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if tc.status == 201 {
				var got Lesson
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.ID != 1 || got.Title != "Go" {
					t.Errorf("lesson=%+v", got)
				}
				if w.Header().Get("Location") != "/lessons/1" || w.Header().Get("Content-Type") != "application/json" {
					t.Errorf("headers=%v", w.Header())
				}
			}
		})
	}
}
