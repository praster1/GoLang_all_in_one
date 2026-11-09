package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"time"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func readValue(ctx context.Context, db *sql.DB) (int, error) {
	var value int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&value)
	return value, err
}

func run() (err error) {
	db, err := openDB(context.Background())
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	held, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	released := false
	defer func() {
		if !released {
			err = errors.Join(err, held.Close())
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, waitErr := readValue(ctx, db)
	fmt.Println("대기 시간 초과:", errors.Is(waitErr, context.DeadlineExceeded))
	if err := held.Close(); err != nil {
		return err
	}
	released = true
	value, err := readValue(context.Background(), db)
	if err != nil {
		return err
	}
	fmt.Println("반환 뒤 조회:", value)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
