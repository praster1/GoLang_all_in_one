package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}
func TestOpenAndCanceledContext(t *testing.T) {
	db := testDB(t)
	var got int
	if err := db.QueryRowContext(context.Background(), "SELECT 1").Scan(&got); err != nil || got != 1 {
		t.Fatalf("got=%d err=%v", got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	failed, err := openDB(ctx)
	if !errors.Is(err, context.Canceled) || failed != nil {
		if failed != nil {
			failed.Close()
		}
		t.Fatalf("db=%v err=%v", failed, err)
	}
	if db.Stats().MaxOpenConnections != 1 {
		t.Fatal("unexpected pool limit")
	}
}
