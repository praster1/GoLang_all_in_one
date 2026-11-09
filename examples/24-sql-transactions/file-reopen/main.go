package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

func openFile(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func save(ctx context.Context, path string) (err error) {
	db, err := openFile(ctx, path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if _, err := db.ExecContext(ctx, "CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT NOT NULL)"); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO notes(title) VALUES (?)", "저장된 Go")
	return err
}

func load(ctx context.Context, path string) (title string, err error) {
	db, err := openFile(ctx, path)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	err = db.QueryRowContext(ctx, "SELECT title FROM notes WHERE id = ?", 1).Scan(&title)
	return title, err
}

func run() (err error) {
	dir, err := os.MkdirTemp("", "go-sql-lesson-")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(dir)) }()
	path := filepath.Join(dir, "lesson.db")
	ctx := context.Background()
	if err := save(ctx, path); err != nil {
		return err
	}
	title, err := load(ctx, path)
	if err != nil {
		return err
	}
	fmt.Println("재열기:", title)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
