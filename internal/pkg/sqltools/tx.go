package sqltools

import (
	"context"
	"database/sql"
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

func Transaction(ctx context.Context, db *sql.DB, f func(ctx context.Context) error) error {
	var err error
	var tx *sql.Tx = new(sql.Tx)

	defer func() {
		if err == nil {
			err = tx.Commit()
		}

		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("error rollback transaction: %w", rbErr)
				return
			}
			err = fmt.Errorf("error commit transaction: %w", err)
			return
		}
	}()

	if _, ok := ctx.Value(txCtxKey{}).(*sql.Tx); !ok {
		tx, err = db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
		})
		if err != nil {
			return fmt.Errorf("error begin transaction: %w", err)
		}

		ctx = context.WithValue(ctx, txCtxKey{}, tx)
	}

	err = f(ctx)
	if err != nil {
		return fmt.Errorf("error run transaction function: %w", err)
	}

	return err
}
