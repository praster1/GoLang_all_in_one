package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestErrorResponses(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		path    string
		status  int
		message string
	}{{"/lessons/x", 400, "invalid_id"}, {"/lessons/0", 400, "invalid_id"}, {"/lessons/-1", 400, "invalid_id"}, {"/lessons/2", 404, "lesson_not_found"}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		var problem Problem
		err := json.Unmarshal(w.Body.Bytes(), &problem)
		if w.Code != tc.status || w.Header().Get("Content-Type") != "application/json" || err != nil || problem.Error != tc.message {
			t.Errorf("path=%s code=%d problem=%+v err=%v", tc.path, w.Code, problem, err)
		}
	}
}
