package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

type storeFunc func(context.Context, int) (Lesson, error)

func (f storeFunc) Lookup(ctx context.Context, id int) (Lesson, error) { return f(ctx, id) }

type markerKey struct{}

func TestLookupOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"found", nil, 200, ""},
		{"missing", ErrNotFound, 404, "not_found"},
		{"wrapped missing", fmt.Errorf("lookup: %w", ErrNotFound), 404, "not_found"},
		{"internal", errors.New("private database detail"), 500, "internal_error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			store := storeFunc(func(ctx context.Context, id int) (Lesson, error) {
				calls++
				if id != 7 || ctx.Value(markerKey{}) != "passed" {
					t.Errorf("id=%d context=%v", id, ctx.Value(markerKey{}))
				}
				return Lesson{ID: 7, Title: "Go"}, tc.err
			})
			request := httptest.NewRequest("GET", "/lessons/7", nil)
			request = request.WithContext(context.WithValue(request.Context(), markerKey{}, "passed"))
			recorder := httptest.NewRecorder()
			newHandler(store).ServeHTTP(recorder, request)
			if recorder.Code != tc.status || calls != 1 {
				t.Fatalf("status=%d calls=%d", recorder.Code, calls)
			}
			if strings.Contains(recorder.Body.String(), "private database detail") {
				t.Fatal("internal error leaked")
			}
			if tc.status == 200 {
				var lesson Lesson
				if err := json.Unmarshal(recorder.Body.Bytes(), &lesson); err != nil || lesson.ID != 7 {
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

func TestInvalidIDDoesNotCallStore(t *testing.T) {
	calls := 0
	store := storeFunc(func(context.Context, int) (Lesson, error) { calls++; return Lesson{}, nil })
	handler := newHandler(store)
	for _, path := range []string{"/lessons/x", "/lessons/0", "/lessons/-1"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
		if recorder.Code != 400 {
			t.Errorf("path=%s status=%d", path, recorder.Code)
		}
	}
	if calls != 0 {
		t.Errorf("invalid requests called store %d times", calls)
	}
}
