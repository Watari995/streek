package database

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type pointLedgerRow struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	HabitID        *string   `db:"habit_id"`
	PointType      string    `db:"type"`
	Amount         int       `db:"amount"`
	Reason         string    `db:"reason"`
	IdempotencyKey string    `db:"idempotency_key"`
	CreatedAt      time.Time `db:"created_at"`
}

type PointLedgerRepository struct {
	db *sqlx.DB
}

func NewPointLedgerRepository(db *sqlx.DB) *PointLedgerRepository {
	return &PointLedgerRepository{db: db}
}

func (r *PointLedgerRepository) Save(ctx context.Context, pointLedger entity.PointLedger) (*entity.PointLedger, error) {
	var habitIDArg any
	if hid := pointLedger.HabitID(); hid != nil {
		habitIDArg = hid.String()
	} else {
		habitIDArg = nil
	}
	conn := getConn(ctx, r.db)
	_, err := conn.ExecContext(ctx, `
		INSERT INTO point_ledger (id, user_id, habit_id, type, amount, reason, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (idempotency_key) DO NOTHING
		`,
		pointLedger.ID().String(), pointLedger.UserID().String(), habitIDArg, pointLedger.PointType().String(), pointLedger.Amount().Int(), pointLedger.Reason().String(), pointLedger.IdempotencyKey(), pointLedger.CreatedAt(),
	)
	if err != nil {
		return nil, err
	}
	return &pointLedger, nil
}


