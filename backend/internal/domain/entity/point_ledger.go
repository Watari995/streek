package entity

import (
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type PointLedger struct {
	id             valueobject.PointLedgerID
	userID         valueobject.UserID
	habitID        *valueobject.HabitID
	pointType      valueobject.PointType
	amount         valueobject.PositiveInt
	reason         valueobject.String50
	idempotencyKey string
	createdAt      time.Time
}

func (p *PointLedger) ID() valueobject.PointLedgerID {
	return p.id
}

func (p *PointLedger) UserID() valueobject.UserID {
	return p.userID
}

func (p *PointLedger) HabitID() *valueobject.HabitID {
	return p.habitID
}

func (p *PointLedger) PointType() valueobject.PointType {
	return p.pointType
}

func (p *PointLedger) Amount() valueobject.PositiveInt {
	return p.amount
}

func (p *PointLedger) Reason() valueobject.String50 {
	return p.reason
}

func (p *PointLedger) IdempotencyKey() string {
	return p.idempotencyKey
}

func (p *PointLedger) CreatedAt() time.Time {
	return p.createdAt
}

func NewPointLedger(
	id valueobject.PointLedgerID,
	userID valueobject.UserID,
	habitID *valueobject.HabitID,
	pointType valueobject.PointType,
	amount valueobject.PositiveInt,
	reason valueobject.String50,
	idempotencyKey string,
	createdAt time.Time,
) PointLedger {
	return PointLedger{
		id:        id,
		userID:    userID,
		habitID:   habitID,
		pointType: pointType,
		amount:    amount,
		reason:    reason,
		idempotencyKey: idempotencyKey,
		createdAt: createdAt,
	}
}

func CreatePointLedger(
	userID valueobject.UserID,
	habitID *valueobject.HabitID,
	pointType valueobject.PointType,
	amount valueobject.PositiveInt,
	reason valueobject.String50,
	idempotencyKey string,
) PointLedger {
	return NewPointLedger(
		valueobject.NewPointLedgerID(),
		userID,
		habitID,
		pointType,
		amount,
		reason,
		idempotencyKey,
		time.Now(),
	)
}
