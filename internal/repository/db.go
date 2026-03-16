package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX abstracts pgxpool.Pool and pgx.Tx for repository operations.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txBeginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

// withTx runs fn inside a transaction when possible. If db does not support
// BeginTx, fn is executed with the existing db (assumed to already be a tx).
func withTx(ctx context.Context, db DBTX, fn func(DBTX) error) error {
	beginner, ok := db.(txBeginner)
	if !ok {
		return fn(db)
	}
	tx, err := beginner.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
