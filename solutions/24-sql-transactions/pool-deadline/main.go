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

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE notes(id INTEGER PRIMARY KEY);INSERT INTO notes VALUES(1),(2)`)
	return err
}
func countNotes(ctx context.Context, db *sql.DB) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count)
	return count, err
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := createSchema(ctx, db); err != nil {
		return err
	}
	held, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	released := false
	defer func() {
		if !released {
			err = errors.Join(err, held.Close())
		}
	}()
	deadline, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, waitErr := countNotes(deadline, db)
	fmt.Println("대기 취소:", errors.Is(waitErr, context.DeadlineExceeded))
	if err := held.Close(); err != nil {
		return err
	}
	released = true
	count, err := countNotes(ctx, db)
	if err != nil {
		return err
	}
	fmt.Println("조회:", count)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
