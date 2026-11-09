package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

const MaxBodyBytes int64 = 1024

type API struct {
	repository Repository
	logger     *slog.Logger
}
type problem struct {
	Error string `json:"error"`
}
type createInput struct {
	Title *string `json:"title"`
}
type replaceInput struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func NewHandler(repository Repository, logger *slog.Logger) (http.Handler, error) {
	if repository == nil {
		return nil, errors.New("repository가 필요합니다")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	api := &API{repository: repository, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("/todos", api.collection)
	mux.HandleFunc("/todos/{id}", api.item)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { api.reply(w, r, 404, problem{"not_found"}) })
	var sequence atomic.Uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A local correlation ID, not a security token or cross-process global ID.
		w.Header().Set("X-Request-ID", strconv.FormatUint(sequence.Add(1), 10))
		mux.ServeHTTP(w, r)
	}), nil
}
func (a *API) reply(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		a.logger.ErrorContext(r.Context(), "encode response", "error", err)
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		a.logger.WarnContext(r.Context(), "write response", "request_id", w.Header().Get("X-Request-ID"), "error", err)
	}
}
func (a *API) failure(w http.ResponseWriter, r *http.Request, err error) {
	var validation *ValidationError
	switch {
	case errors.As(err, &validation):
		a.reply(w, r, 422, problem{"invalid_" + validation.Field})
	case errors.Is(err, ErrInvalidID), errors.Is(err, ErrInvalidPage):
		a.reply(w, r, 400, problem{"invalid_parameter"})
	case errors.Is(err, ErrNotFound):
		a.reply(w, r, 404, problem{"not_found"})
	case errors.Is(err, context.DeadlineExceeded):
		a.reply(w, r, 504, problem{"deadline_exceeded"})
	case errors.Is(err, context.Canceled):
		a.reply(w, r, 503, problem{"request_canceled"})
	default:
		a.logger.ErrorContext(r.Context(), "repository failure", "request_id", w.Header().Get("X-Request-ID"), "error", err)
		a.reply(w, r, 500, problem{"internal_error"})
	}
}
func (a *API) decode(w http.ResponseWriter, r *http.Request, target any) bool {
	media, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" || (params["charset"] != "" && !strings.EqualFold(params["charset"], "utf-8")) {
		a.reply(w, r, 415, problem{"json_required"})
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var max *http.MaxBytesError
		if errors.As(err, &max) {
			a.reply(w, r, 413, problem{"body_too_large"})
		} else {
			a.reply(w, r, 400, problem{"invalid_body"})
		}
		return false
	}
	if !utf8.Valid(body) {
		a.reply(w, r, 400, problem{"invalid_json"})
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		a.reply(w, r, 400, problem{"invalid_json"})
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		a.reply(w, r, 400, problem{"invalid_json"})
		return false
	}
	return true
}
func singleQuery(values url.Values, key string) (string, bool) {
	all, exists := values[key]
	if !exists {
		return "", true
	}
	if len(all) != 1 || all[0] == "" {
		return "", false
	}
	return all[0], true
}
func page(r *http.Request) (int64, int, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return 0, 0, ErrInvalidPage
	}
	for key := range values {
		if key != "after" && key != "limit" {
			return 0, 0, ErrInvalidPage
		}
	}
	afterText, ok := singleQuery(values, "after")
	if !ok {
		return 0, 0, ErrInvalidPage
	}
	limitText, ok := singleQuery(values, "limit")
	if !ok {
		return 0, 0, ErrInvalidPage
	}
	var after int64
	limit := 20
	if afterText != "" {
		after, err = strconv.ParseInt(afterText, 10, 64)
		if err != nil {
			return 0, 0, ErrInvalidPage
		}
	}
	if limitText != "" {
		limit, err = strconv.Atoi(limitText)
		if err != nil {
			return 0, 0, ErrInvalidPage
		}
	}
	if err := validatePage(after, limit); err != nil {
		return 0, 0, err
	}
	return after, limit, nil
}
func (a *API) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		after, limit, err := page(r)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		tasks, err := a.repository.List(r.Context(), after, limit)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		a.reply(w, r, 200, tasks)
	case http.MethodPost:
		if r.URL.RawQuery != "" {
			a.reply(w, r, 400, problem{"invalid_parameter"})
			return
		}
		var input *createInput
		if !a.decode(w, r, &input) {
			return
		}
		if input == nil || input.Title == nil {
			a.reply(w, r, 422, problem{"invalid_title"})
			return
		}
		task, err := a.repository.Create(r.Context(), *input.Title)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		w.Header().Set("Location", "/todos/"+strconv.FormatInt(task.ID, 10))
		a.reply(w, r, 201, task)
	default:
		w.Header().Set("Allow", "GET, HEAD, POST")
		a.reply(w, r, 405, problem{"method_not_allowed"})
	}
}
func (a *API) item(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodPut && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "GET, HEAD, PUT, DELETE")
		a.reply(w, r, 405, problem{"method_not_allowed"})
		return
	}
	if r.URL.RawQuery != "" {
		a.reply(w, r, 400, problem{"invalid_parameter"})
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		a.failure(w, r, ErrInvalidID)
		return
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		task, err := a.repository.Get(r.Context(), id)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		a.reply(w, r, 200, task)
	case http.MethodPut:
		var input *replaceInput
		if !a.decode(w, r, &input) {
			return
		}
		if input == nil || input.Title == nil || input.Done == nil {
			a.reply(w, r, 422, problem{"missing_field"})
			return
		}
		task, err := a.repository.Replace(r.Context(), id, *input.Title, *input.Done)
		if err != nil {
			a.failure(w, r, err)
			return
		}
		a.reply(w, r, 200, task)
	case http.MethodDelete:
		if err := a.repository.Delete(r.Context(), id); err != nil {
			a.failure(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
