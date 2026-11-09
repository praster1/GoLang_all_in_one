package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestPathValues(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		path       string
		status, id int
	}{{"/lessons/1", 200, 1}, {"/lessons/7", 200, 7}, {"/lessons/0", 400, 0}, {"/lessons/-1", 400, 0}, {"/lessons/x", 400, 0}, {"/lessons/9999999999999999999999999", 400, 0}, {"/lessons/7/extra", 404, 0}} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.status {
			t.Errorf("path=%s code=%d", tc.path, w.Code)
		}
		if tc.status == 200 {
			var lesson Lesson
			if err := json.Unmarshal(w.Body.Bytes(), &lesson); err != nil || lesson.ID != tc.id || lesson.Title != "Go" {
				t.Errorf("lesson=%+v err=%v", lesson, err)
			}
		}
	}
}
