package database

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
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

func (r *PointLedgerRepository) GetBalance(ctx context.Context, userID valueobject.UserID) (int, error) {
	conn := getConn(ctx, r.db)
	var balance int
	err := conn.QueryRowxContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN type = 'EARN' THEN amount ELSE 0 END), 0) -
		COALESCE(SUM(CASE WHEN type = 'SPEND' THEN amount ELSE 0 END), 0)
		AS balance
		FROM point_ledger
		WHERE user_id = $1
		`, userID.String()).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (r *PointLedgerRepository) FindByUserID(ctx context.Context, userID valueobject.UserID) ([]*entity.PointLedger, error) {
	rows := []pointLedgerRow{}
	conn := getConn(ctx, r.db)
	err := sqlx.SelectContext(ctx, conn, &rows, `
		SELECT id, user_id, habit_id, type, amount, reason, idempotency_key, created_at
		FROM point_ledger
		WHERE user_id = $1
		ORDER BY created_at DESC
		`, userID.String())
	if err != nil {
		return nil, err
	}
	pointLedgers := make([]*entity.PointLedger, len(rows))
	for i, row := range rows {
		pointLedger, err := r.toEntity(row)
		if err != nil {
			return nil, err
		}
		pointLedgers[i] = pointLedger
	}
	return pointLedgers, nil
}

func (r *PointLedgerRepository) ExistsByIdempotencyKey(ctx context.Context, idempotencyKey string) (bool, error) {
	conn := getConn(ctx, r.db)
	var exists bool
	err := conn.QueryRowxContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM point_ledger WHERE idempotency_key = $1)
		`, idempotencyKey).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// private method
func (r *PointLedgerRepository) toEntity(row pointLedgerRow) (*entity.PointLedger, error) {
	pointLedgerID, err := valueobject.NewPointLedgerIDFromString(row.ID)
	if err != nil {
		return nil, err
	}
	userID, err := valueobject.NewUserIDFromString(row.UserID)
	if err != nil {
		return nil, err
	}
	var habitID *valueobject.HabitID
	if row.HabitID != nil {
		habitIDValue, err := valueobject.NewHabitIDFromString(*row.HabitID)
		if err != nil {
			return nil, err
		}
		habitID = &habitIDValue
	}
	pointType, err := valueobject.NewPointType(row.PointType)
	if err != nil {
		return nil, err
	}
	amount, err := valueobject.NewPositiveInt(row.Amount)
	if err != nil {
		return nil, err
	}
	reason, err := valueobject.NewString50(row.Reason)
	if err != nil {
		return nil, err
	}

	pointLedger := entity.NewPointLedger(
		pointLedgerID,
		userID,
		habitID,
		pointType,
		amount,
		reason,
		row.IdempotencyKey,
		row.CreatedAt,
	)
	return &pointLedger, nil
}
