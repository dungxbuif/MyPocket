// Package commandtx shares a SQL transaction between domain commands and sync.
package commandtx

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
)

type Queryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Handle struct {
	Queryer
	pool *sql.DB
	tx   *sql.Tx
}

func New(db *sql.DB) *Handle   { return &Handle{Queryer: db, pool: db} }
func Bound(tx *sql.Tx) *Handle { return &Handle{Queryer: tx, tx: tx} }

var sequence atomic.Uint64

type Scope struct {
	*sql.Tx
	ctx       context.Context
	savepoint string
	done      bool
}

func (h *Handle) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Scope, error) {
	if h.tx == nil {
		tx, err := h.pool.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &Scope{Tx: tx, ctx: ctx}, nil
	}
	name := fmt.Sprintf("command_%d", sequence.Add(1))
	if _, err := h.tx.ExecContext(ctx, "SAVEPOINT "+name); err != nil {
		return nil, err
	}
	return &Scope{Tx: h.tx, ctx: ctx, savepoint: name}, nil
}
func (s *Scope) Commit() error {
	if s.done {
		return sql.ErrTxDone
	}
	s.done = true
	if s.savepoint == "" {
		return s.Tx.Commit()
	}
	_, err := s.Tx.ExecContext(s.ctx, "RELEASE SAVEPOINT "+s.savepoint)
	return err
}
func (s *Scope) Rollback() error {
	if s.done {
		return sql.ErrTxDone
	}
	s.done = true
	if s.savepoint == "" {
		return s.Tx.Rollback()
	}
	if _, err := s.Tx.ExecContext(s.ctx, "ROLLBACK TO SAVEPOINT "+s.savepoint); err != nil {
		return err
	}
	_, err := s.Tx.ExecContext(s.ctx, "RELEASE SAVEPOINT "+s.savepoint)
	return err
}

// LockUser orders financial writers before entity locks and cursor reservation.
// It also protects receipt lookup when the mutation ID has no row yet.
func LockUser(ctx context.Context, q Queryer, userID string) error {
	_, err := q.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 803))`, userID)
	return err
}

func Run[T any](ctx context.Context, h *Handle, userID string, fn func(*Handle) (T, error)) (T, error) {
	var zero T
	tx, err := h.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	if err = LockUser(ctx, tx, userID); err != nil {
		return zero, err
	}
	value, err := fn(Bound(tx.Tx))
	if err != nil {
		return zero, err
	}
	if err = tx.Commit(); err != nil {
		return zero, err
	}
	return value, nil
}
