package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func request(t *testing.T, handler http.Handler, method, path, body, media string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if media != "" {
		r.Header.Set("Content-Type", media)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}
func TestHTTPContractBothRepositories(t *testing.T) {
	for name, factory := range repositories(t) {
		t.Run(name, func(t *testing.T) {
			handler, err := NewHandler(factory(t), nil)
			if err != nil {
				t.Fatal(err)
			}
			initial := request(t, handler, "GET", "/todos", "", "")
			if initial.Code != 200 || initial.Body.String() != "[]\n" {
				t.Fatalf("initial=%d %s", initial.Code, initial.Body.String())
			}
			created := request(t, handler, "POST", "/todos", `{"title":" Go "}`, "application/json")
			if created.Code != 201 || created.Header().Get("Location") != "/todos/1" {
				t.Fatalf("create=%d %s", created.Code, created.Body.String())
			}
			var task Task
			if err := json.Unmarshal(created.Body.Bytes(), &task); err != nil {
				t.Fatal(err)
			}
			if task != (Task{ID: 1, Title: "Go"}) {
				t.Fatalf("task=%+v", task)
			}
			got := request(t, handler, "GET", "/todos/1", "", "")
			if got.Code != 200 || got.Body.String() != created.Body.String() {
				t.Fatalf("get=%d %s", got.Code, got.Body.String())
			}
			replaced := request(t, handler, "PUT", "/todos/1", `{"title":"복습","done":false}`, "application/json; charset=utf-8")
			if replaced.Code != 200 || !strings.Contains(replaced.Body.String(), `"done":false`) {
				t.Fatalf("replace=%d %s", replaced.Code, replaced.Body.String())
			}
			head := request(t, handler, "HEAD", "/todos/1", "", "")
			if head.Code != 200 || head.Body.Len() != 0 {
				t.Fatal("HEAD has body")
			}
			deleted := request(t, handler, "DELETE", "/todos/1", "", "")
			if deleted.Code != 204 || deleted.Body.Len() != 0 {
				t.Fatal("DELETE not empty")
			}
			missing := request(t, handler, "GET", "/todos/1", "", "")
			if missing.Code != 404 || missing.Body.String() != "{\"error\":\"not_found\"}\n" {
				t.Fatal("missing contract")
			}
		})
	}
}
func TestHTTPInputFailures(t *testing.T) {
	for name, factory := range repositories(t) {
		t.Run(name, func(t *testing.T) {
			repo := factory(t)
			handler, err := NewHandler(repo, nil)
			if err != nil {
				t.Fatal(err)
			}
			cases := []struct {
				name, method, path, body, media string
				status                          int
			}{
				{"missing-media", "POST", "/todos", `{"title":"x"}`, "", 415},
				{"wrong-media", "POST", "/todos", `{"title":"x"}`, "text/plain", 415},
				{"wrong-charset", "POST", "/todos", `{"title":"x"}`, "application/json; charset=latin1", 415},
				{"bad-json", "POST", "/todos", `{"title":`, "application/json", 400},
				{"extra-value", "POST", "/todos", `{"title":"x"} {}`, "application/json", 400},
				{"unknown-field", "POST", "/todos", `{"title":"x","admin":true}`, "application/json", 400},
				{"array", "POST", "/todos", `[]`, "application/json", 400},
				{"null", "POST", "/todos", `null`, "application/json", 422},
				{"missing-title", "POST", "/todos", `{}`, "application/json", 422},
				{"null-title", "POST", "/todos", `{"title":null}`, "application/json", 422},
				{"empty-title", "POST", "/todos", `{"title":" "}`, "application/json", 422},
				{"long-title", "POST", "/todos", `{"title":"` + strings.Repeat("가", 81) + `"}`, "application/json", 422},
				{"invalid-utf8", "POST", "/todos", "{\"title\":\"" + string([]byte{0xff}) + "\"}", "application/json", 400},
				{"body-limit", "POST", "/todos", `{"title":"x"}` + strings.Repeat(" ", 1024), "application/json", 413},
				{"bad-id", "GET", "/todos/abc", "", "", 400},
				{"zero-id", "GET", "/todos/0", "", "", 400},
				{"invalid-limit", "GET", "/todos?limit=101", "", "", 400},
				{"duplicate-limit", "GET", "/todos?limit=1&limit=2", "", "", 400},
				{"empty-limit", "GET", "/todos?limit=", "", "", 400},
				{"negative-after", "GET", "/todos?after=-1", "", "", 400},
				{"unknown-query", "GET", "/todos?secret=x", "", "", 400},
				{"malformed-query", "GET", "/todos?limit=%XX", "", "", 400},
				{"missing-done", "PUT", "/todos/1", `{"title":"x"}`, "application/json", 422},
				{"null-done", "PUT", "/todos/1", `{"title":"x","done":null}`, "application/json", 422},
				{"method", "PATCH", "/todos/1", `{}`, "application/json", 405},
				{"collection-method", "DELETE", "/todos", "", "", 405},
				{"unknown-route", "GET", "/unknown", "", "", 404},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					w := request(t, handler, tc.method, tc.path, tc.body, tc.media)
					if w.Code != tc.status {
						t.Fatalf("status=%d want=%d body=%s", w.Code, tc.status, w.Body.String())
					}
					var p problem
					if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil || p.Error == "" {
						t.Fatalf("not JSON problem: %s error=%v", w.Body.String(), err)
					}
					if w.Header().Get("X-Request-ID") == "" {
						t.Fatal("missing request id")
					}
					if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
						t.Fatal("wrong media")
					}
					if tc.status == 405 && w.Header().Get("Allow") == "" {
						t.Fatal("missing Allow")
					}
				})
			}
			tasks, err := repo.List(context.Background(), 0, 20)
			if err != nil || len(tasks) != 0 {
				t.Fatalf("invalid requests changed state: %v %v", tasks, err)
			}
		})
	}
}
func TestExactlyBoundedBody(t *testing.T) {
	handler, err := NewHandler(NewMemory(), nil)
	if err != nil {
		t.Fatal(err)
	}
	body := `{"title":"x"}`
	body += strings.Repeat(" ", int(MaxBodyBytes)-len(body))
	if w := request(t, handler, "POST", "/todos", body, "application/json"); w.Code != 201 {
		t.Fatalf("limit body=%d %s", w.Code, w.Body.String())
	}
	if w := request(t, handler, "POST", "/todos", body+" ", "application/json"); w.Code != 413 {
		t.Fatalf("limit+1=%d", w.Code)
	}
}

type failedRepository struct {
	Repository
	err   error
	calls int
	ctx   context.Context
}

func (f *failedRepository) Get(ctx context.Context, id int64) (Task, error) {
	f.calls++
	f.ctx = ctx
	return Task{}, f.err
}
func TestRepositoryErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{fmt.Errorf("wrapped: %w", ErrNotFound), 404, "not_found"},
		{context.DeadlineExceeded, 504, "deadline_exceeded"},
		{context.Canceled, 503, "request_canceled"},
		{errors.New("private-storage-path"), 500, "internal_error"},
	} {
		repo := &failedRepository{err: tc.err}
		var logs bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logs, nil))
		handler, err := NewHandler(repo, logger)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		r := httptest.NewRequest("GET", "/todos/1", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.code) || strings.Contains(w.Body.String(), "private-storage-path") {
			t.Fatalf("response=%d %s", w.Code, w.Body.String())
		}
		if repo.ctx != ctx || repo.calls != 1 {
			t.Fatal("lost context or duplicate call")
		}
		before := repo.calls
		bad := request(t, handler, "GET", "/todos/nope", "", "")
		if bad.Code != 400 || repo.calls != before {
			t.Fatal("invalid ID reached repository")
		}
	}
}
func TestJSONV1DuplicateAndCasePolicy(t *testing.T) {
	handler, err := NewHandler(NewMemory(), nil)
	if err != nil {
		t.Fatal(err)
	}
	w := request(t, handler, "POST", "/todos", `{"TITLE":"first","title":"second"}`, "application/json")
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"title":"second"`) {
		t.Fatalf("documented JSON v1 behavior changed: %d %s", w.Code, w.Body.String())
	}
}
func TestRequestIDsConcurrent(t *testing.T) {
	handler, err := NewHandler(NewMemory(), nil)
	if err != nil {
		t.Fatal(err)
	}
	const count = 40
	var wg sync.WaitGroup
	ids := make(chan string, count)
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", "/todos", nil))
			ids <- w.Header().Get("X-Request-ID")
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[string]bool)
	for id := range ids {
		if id == "" || seen[id] {
			t.Fatalf("duplicate ID %q", id)
		}
		seen[id] = true
	}
}
func TestRealHTTP(t *testing.T) {
	handler, err := NewHandler(NewMemory(), nil)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	client := server.Client()
	defer client.CloseIdleConnections()
	response, err := client.Post(server.URL+"/todos", "application/json", strings.NewReader(`{"title":"TLS"}`))
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 201 || !strings.Contains(string(body), "TLS") {
		t.Fatalf("post=%d %s", response.StatusCode, body)
	}
	request, err := http.NewRequest("HEAD", server.URL+"/todos/1", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, readErr = io.ReadAll(response.Body)
	closeErr = response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || len(body) != 0 {
		t.Fatalf("HEAD=%d %q", response.StatusCode, body)
	}
}
