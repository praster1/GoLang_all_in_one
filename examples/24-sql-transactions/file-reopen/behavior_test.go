package main

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFileReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lesson.db")
	ctx := context.Background()
	if err := save(ctx, path); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		got, err := load(ctx, path)
		if err != nil || got != "저장된 Go" {
			t.Fatalf("got=%q err=%v", got, err)
		}
	}
}
func TestUnavailableDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "lesson.db")
	db, err := openFile(context.Background(), path)
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("expected open failure")
	}
	if db != nil {
		t.Fatal("failed open returned database")
	}
}
