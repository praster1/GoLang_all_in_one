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

var ErrSource = errors.New("source missing or points insufficient")
var ErrDestination = errors.New("destination missing")
var ErrInput = errors.New("invalid transfer input")

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE wallets (id INTEGER PRIMARY KEY, points INTEGER NOT NULL CHECK(points >= 0));
        INSERT INTO wallets VALUES (1, 50), (2, 10)`)
	return err
}

func transfer(ctx context.Context, db *sql.DB, from, to int64, amount int) (err error) {
	if from <= 0 || to <= 0 || from == to || amount <= 0 || amount > 1000 {
		return ErrInput
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
	result, err := tx.ExecContext(ctx, "UPDATE wallets SET points = points - ? WHERE id = ? AND points >= ?", amount, from, amount)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrSource
	}
	result, err = tx.ExecContext(ctx, "UPDATE wallets SET points = points + ? WHERE id = ?", amount, to)
	if err != nil {
		return err
	}
	changed, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrDestination
	}
	return tx.Commit()
}

func points(ctx context.Context, db *sql.DB, id int64) (int, error) {
	var value int
	err := db.QueryRowContext(ctx, "SELECT points FROM wallets WHERE id = ?", id).Scan(&value)
	return value, err
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
	if err := transfer(ctx, db, 1, 2, 15); err != nil {
		return err
	}
	first, err := points(ctx, db, 1)
	if err != nil {
		return err
	}
	second, err := points(ctx, db, 2)
	if err != nil {
		return err
	}
	fmt.Printf("성공: %d, %d\n", first, second)
	moveErr := transfer(ctx, db, 1, 9, 5)
	fmt.Println("복구 필요:", errors.Is(moveErr, ErrDestination))
	first, err = points(ctx, db, 1)
	if err != nil {
		return err
	}
	second, err = points(ctx, db, 2)
	if err != nil {
		return err
	}
	fmt.Printf("유지: %d, %d\n", first, second)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
