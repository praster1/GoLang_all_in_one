package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
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

type Note struct {
	ID    int64
	Title sql.NullString
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE notes(id INTEGER PRIMARY KEY,title TEXT);INSERT INTO notes VALUES(1,'Go'),(2,NULL),(3,'')`)
	return err
}
func noteLabel(ctx context.Context, db *sql.DB, id int64) (string, error) {
	var title sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT title FROM notes WHERE id = ?", id).Scan(&title); err != nil {
		return "", fmt.Errorf("note label: %w", err)
	}
	if !title.Valid {
		return "(미정)", nil
	}
	if title.String == "" {
		return "(빈 문자열)", nil
	}
	return title.String, nil
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
	for _, id := range []int64{1, 2, 3} {
		label, err := noteLabel(ctx, db, id)
		if err != nil {
			return err
		}
		fmt.Printf("%d: %s\n", id, label)
	}
	_, missing := noteLabel(ctx, db, 99)
	fmt.Println("행 없음:", errors.Is(missing, sql.ErrNoRows))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
