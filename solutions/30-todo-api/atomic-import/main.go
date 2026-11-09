package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
	_ "modernc.org/sqlite"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.ExecContext(ctx, "CREATE TABLE imports(id INTEGER PRIMARY KEY,title TEXT NOT NULL UNIQUE)"); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}
func importTasks(ctx context.Context, db *sql.DB, titles []string) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(titles) < 1 || len(titles) > 100 {
		return errors.New("항목은 1~100개여야 합니다")
	}
	normalized := make([]string, len(titles))
	for i, title := range titles {
		value, err := todo.NormalizeTitle(title)
		if err != nil {
			return fmt.Errorf("항목 %d: %w", i+1, err)
		}
		normalized[i] = value
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, rollbackErr)
		}
	}()
	for _, title := range normalized {
		if _, err := tx.ExecContext(ctx, "INSERT INTO imports(title) VALUES (?)", title); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func count(ctx context.Context, db *sql.DB) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, "SELECT count(*) FROM imports").Scan(&n)
	return n, err
}
func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := importTasks(ctx, db, []string{"기존"}); err != nil {
		return err
	}
	duplicateErr := importTasks(ctx, db, []string{"새 항목", "기존"})
	fmt.Println("중복 실패:", duplicateErr != nil)
	n, err := count(ctx, db)
	if err != nil {
		return err
	}
	fmt.Println("실패 뒤 개수:", n)
	if err := importTasks(ctx, db, []string{"Go", "SQL"}); err != nil {
		return err
	}
	n, err = count(ctx, db)
	if err != nil {
		return err
	}
	fmt.Println("성공 뒤 개수:", n)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
