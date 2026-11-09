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

type Lesson struct {
	Title   string `json:"title"`
	Minutes int    `json:"minutes"`
}

func decodeLesson(reader io.Reader) (Lesson, error) {
	var lesson Lesson
	body, err := io.ReadAll(io.LimitReader(reader, 257))
	if err != nil {
		return lesson, err
	}
	if len(body) > 256 {
		return lesson, errors.New("본문 한도 초과")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&lesson); err != nil {
		return Lesson{}, fmt.Errorf("JSON 해석: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Lesson{}, errors.New("JSON 값은 하나여야 함")
	}
	if strings.TrimSpace(lesson.Title) == "" || lesson.Minutes < 1 || lesson.Minutes > 180 {
		return Lesson{}, errors.New("학습 데이터 범위 오류")
	}
	return lesson, nil
}

func fetchLesson(ctx context.Context, client *http.Client, address string) (Lesson, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return Lesson{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Lesson{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Lesson{}, fmt.Errorf("상태: %d", resp.StatusCode)
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return Lesson{}, errors.New("JSON Content-Type 필요")
	}
	return decodeLesson(resp.Body)
}

func run() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(Lesson{Title: "함수", Minutes: 45}); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	lesson, err := fetchLesson(context.Background(), client, server.URL)
	if err != nil {
		return err
	}
	fmt.Printf("학습: %s, %d분\n", lesson.Title, lesson.Minutes)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
