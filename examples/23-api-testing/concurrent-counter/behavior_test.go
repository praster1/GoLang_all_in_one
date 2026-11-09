package main

import (
	"encoding/json"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestConcurrentIncrements(t *testing.T) {
	const count = 50
	handler := newHandler()
	results := make(chan int64, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest("POST", "/visits", nil))
			if recorder.Code != 200 {
				t.Errorf("status=%d", recorder.Code)
				return
			}
			var body struct {
				Count int64 `json:"count"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Error(err)
				return
			}
			results <- body.Count
		}()
	}
	wg.Wait()
	close(results)
	seen := make(map[int64]bool)
	for value := range results {
		if value < 1 || value > count || seen[value] {
			t.Errorf("invalid or duplicate count=%d", value)
		}
		seen[value] = true
	}
	if len(seen) != count {
		t.Fatalf("unique values=%d, want %d", len(seen), count)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/visits", nil))
	var body struct {
		Count int64 `json:"count"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.Count != count {
		t.Errorf("final=%d err=%v", body.Count, err)
	}
}
func TestReadDoesNotIncrement(t *testing.T) {
	handler := newHandler()
	for range 2 {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/visits", nil))
		if recorder.Body.String() != "{\"count\":0}\n" {
			t.Errorf("body=%q", recorder.Body.String())
		}
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("DELETE", "/visits", nil))
	if recorder.Code != 405 {
		t.Errorf("DELETE status=%d", recorder.Code)
	}
}
