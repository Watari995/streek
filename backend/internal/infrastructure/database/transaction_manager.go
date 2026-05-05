package database

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type txKey struct{}

const (
	deadlockDetectedCode = "40P01"
	maxRetryCount        = 3
)

type TransactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetTx(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx, ok
}

func (m *TransactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	var lastErr error
	for i := 0; i < maxRetryCount; i++ {
		err := m.runOnce(ctx, fn)
		if err == nil {
			return nil
		}
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == deadlockDetectedCode {
			lastErr = err
			continue // retry for deadlock
		}
		return err
	}
	return lastErr
}

func (m *TransactionManager) runOnce(ctx context.Context, fn func(ctx context.Context) error) error {
	// if already in tx, execute fn directly
	if _, ok := GetTx(ctx); ok {
		return fn(ctx)
	}
	// start new tx
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			// after rollback, panic again
			panic(r)
		}
	}()
	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func getConn(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx, ok := GetTx(ctx); ok {
		return tx
	}
	return db
}
