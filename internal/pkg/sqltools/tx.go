package sqltools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type DBTX interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type TransactionalStorage interface {
	Transaction(ctx context.Context, db *sql.DB, f func(ctx context.Context) error) error
	Conn(ctx context.Context) DBTX
}

type txCtxKey struct{}

func Transaction(ctx context.Context, db *sql.DB, fn func(context.Context) error) error {
	var err error

	var tx *sql.Tx = new(sql.Tx)

	hasExternalTx := hasExternalTransaction(ctx)

	defer func() {
		if hasExternalTx {
			if err != nil {
				err = fmt.Errorf("error perform operation. %w", err)
				return
			}

			return
		}

		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = errors.Join(fmt.Errorf("error rollback transaction. %w", rbErr), err)
				return
			}

			err = fmt.Errorf("error execute transactional operation. %w", err)

			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = errors.Join(fmt.Errorf("error rollback transaction. %w", rbErr), commitErr, err)

				return
			}

			err = fmt.Errorf("error commit transaction. %w", err)
		}
	}()

	if !hasExternalTx {
		tx, err = db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
		})
		if err != nil {
			return fmt.Errorf("error begin transaction. %w", err)
		}

		ctx = context.WithValue(ctx, txCtxKey{}, tx)
	}

	return fn(ctx)
}

func hasExternalTransaction(ctx context.Context) bool {
	if _, ok := ctx.Value(txCtxKey{}).(*sql.Tx); ok {
		return true
	}

	return false
}
