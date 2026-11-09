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
	if err := createSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestAtomicTransfer(t *testing.T) {
	for _, tc := range []struct {
		name          string
		from, to      int64
		amount        int
		wantErr       error
		first, second int
	}{
		{"normal", 1, 2, 15, nil, 35, 25}, {"empty source after move", 1, 2, 50, nil, 0, 60}, {"insufficient", 1, 2, 51, ErrSource, 50, 10},
		{"missing destination", 1, 9, 5, ErrDestination, 50, 10}, {"missing source", 9, 2, 5, ErrSource, 50, 10},
		{"same", 1, 1, 5, ErrInput, 50, 10}, {"zero", 1, 2, 0, ErrInput, 50, 10}, {"negative", 1, 2, -1, ErrInput, 50, 10},
		{"limit", 1, 2, 1001, ErrInput, 50, 10}, {"invalid id", 0, 2, 1, ErrInput, 50, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := testDB(t)
			ctx := context.Background()
			err := transfer(ctx, db, tc.from, tc.to, tc.amount)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err=%v", err)
			}
			first, err := points(ctx, db, 1)
			if err != nil {
				t.Fatal(err)
			}
			second, err := points(ctx, db, 2)
			if err != nil {
				t.Fatal(err)
			}
			if first != tc.first || second != tc.second || first+second != 60 {
				t.Fatalf("points=%d,%d", first, second)
			}
		})
	}
}
func TestCanceledTransfer(t *testing.T) {
	db := testDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := transfer(ctx, db, 1, 2, 15); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	got, err := points(context.Background(), db, 1)
	if err != nil || got != 50 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}
