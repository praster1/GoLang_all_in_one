package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

type PlanInput struct {
	Days    int `json:"days"`
	Minutes int `json:"minutes"`
}

type PlanPreview struct {
	TotalMinutes int `json:"total_minutes"`
}

func readInput(w http.ResponseWriter, r *http.Request) (PlanInput, error) {
	var input PlanInput
	r.Body = http.MaxBytesReader(w, r.Body, 128)
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return input, err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return input, fmt.Errorf("JSON 값은 하나여야 함")
	}
	return input, nil
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /plans/preview", func(w http.ResponseWriter, r *http.Request) {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			http.Error(w, "JSON 형식이 필요합니다", http.StatusUnsupportedMediaType)
			return
		}
		input, err := readInput(w, r)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				http.Error(w, "본문이 너무 큽니다", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "잘못된 JSON입니다", http.StatusBadRequest)
			}
			return
		}
		if input.Days < 0 || input.Days > 7 || input.Minutes < 1 || input.Minutes > 120 {
			http.Error(w, "계획 값을 확인하세요", http.StatusUnprocessableEntity)
			return
		}
		data, err := json.Marshal(PlanPreview{TotalMinutes: input.Days * input.Minutes})
		if err != nil {
			http.Error(w, "응답 생성 실패", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(append(data, '\n')); err != nil {
			return
		}
	})
	return mux
}

func call(client *http.Client, address, method, body string) (int, string, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, address, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil {
		return 0, "", err
	}
	if len(data) > 4096 {
		return 0, "", fmt.Errorf("응답 한도 초과")
	}
	return resp.StatusCode, string(data), nil
}

func run() error {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/plans/preview", `{"days":3,"minutes":25}`},
		{http.MethodPost, "/plans/preview", `{"days":0,"minutes":25}`},
		{http.MethodPost, "/plans/preview", `{"days":8,"minutes":25}`},
	} {
		status, text, err := call(client, server.URL+tc.path, tc.method, tc.body)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s -> %d: %s\n", tc.method, tc.path, status, strings.TrimSpace(text))
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
