package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// Handle is the first backend-shell seam for database access.
// The first concrete implementation uses database/sql so repositories can talk to
// the real Postgres schema without forcing the rest of the backend to know about
// driver details.
type Handle struct {
	dsn     string
	db      *sql.DB
	openErr error
}

func New(dsn string) *Handle {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return &Handle{dsn: dsn}
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return &Handle{
			dsn:     dsn,
			openErr: fmt.Errorf("open database: %w", err),
		}
	}

	return &Handle{
		dsn: dsn,
		db:  db,
	}
}

func (h *Handle) Ping(ctx context.Context) error {
	if h == nil || h.dsn == "" {
		return errors.New("database is not configured")
	}

	if h.openErr != nil {
		return h.openErr
	}

	if h.db == nil {
		return errors.New("database is not initialized")
	}

	return h.db.PingContext(ctx)
}

func (h *Handle) DB() DBTX {
	if h == nil {
		return nil
	}

	return h.db
}

func (h *Handle) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if h == nil || h.dsn == "" {
		return nil, errors.New("database is not configured")
	}
	if h.openErr != nil {
		return nil, h.openErr
	}
	if h.db == nil {
		return nil, errors.New("database is not initialized")
	}

	return h.db.BeginTx(ctx, opts)
}

func (h *Handle) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	if fn == nil {
		return errors.New("transaction callback is required")
	}

	tx, err := h.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			return fmt.Errorf("rollback transaction: %v after %w", rollbackErr, err)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (h *Handle) Close() error {
	if h == nil || h.db == nil {
		return nil
	}

	return h.db.Close()
}
