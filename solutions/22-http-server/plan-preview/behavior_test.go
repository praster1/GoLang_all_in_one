package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPreviewContract(t *testing.T) {
	valid := `{"days":3,"minutes":25}`
	for _, tc := range []struct {
		name, body, media string
		status, total     int
	}{
		{"normal", valid, "application/json", 200, 75},
		{"omitted days default to zero", `{"minutes":25}`, "application/json", 200, 0}, {"zero days", `{"days":0,"minutes":25}`, "application/json", 200, 0},
		{"minimum", `{"days":1,"minutes":1}`, "application/json", 200, 1}, {"maximum", `{"days":7,"minutes":120}`, "application/json", 200, 840},
		{"negative", `{"days":-1,"minutes":25}`, "application/json", 422, 0}, {"days above", `{"days":8,"minutes":25}`, "application/json", 422, 0},
		{"minutes zero", `{"days":0,"minutes":0}`, "application/json", 422, 0}, {"minutes above", `{"days":5,"minutes":121}`, "application/json", 422, 0},
		{"unknown", `{"days":3,"minutes":25,"extra":1}`, "application/json", 400, 0}, {"two values", valid + "{}", "application/json", 400, 0},
		{"trailing junk", valid + "x", "application/json", 400, 0}, {"empty", "", "application/json", 400, 0}, {"null", "null", "application/json", 422, 0},
		{"exact limit", valid + strings.Repeat(" ", 128-len(valid)), "application/json", 200, 75},
		{"too large", valid + strings.Repeat(" ", 129-len(valid)), "application/json", 413, 0}, {"wrong media", valid, "text/plain", 415, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/plans/preview", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.media)
			w := httptest.NewRecorder()
			newHandler().ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("code=%d body=%q", w.Code, w.Body.String())
			}
			if tc.status == 200 {
				var result PlanPreview
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.TotalMinutes != tc.total {
					t.Errorf("result=%+v err=%v", result, err)
				}
			}
			if w.Header().Get("Location") != "" {
				t.Error("preview must not advertise stored resource")
			}
		})
	}
}
