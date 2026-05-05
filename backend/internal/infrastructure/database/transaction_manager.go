package database

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

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
	// 既に tx に参加している場合は、fn をそのまま実行
	if _, ok := GetTx(ctx); ok {
		return fn(ctx)
	}
	// 新規 tx を開始
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
