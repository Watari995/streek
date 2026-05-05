package types

import (
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

const (
	EventTypeCheckInCompleted = "checkInCompleted"
)

type CheckInCompletedEvent struct {
	UserID         valueobject.UserID
	HabitID        valueobject.HabitID
	CheckedDate    valueobject.DateString
	PointAmount    valueobject.PositiveInt
	PointReason    valueobject.String50
	IdempotencyKey string
	CreatedAt      time.Time
}

func NewCheckInCompletedEvent(
	userID valueobject.UserID,
	habitID valueobject.HabitID,
	checkedDate valueobject.DateString,
	pointAmount valueobject.PositiveInt,
	pointReason valueobject.String50,
	idempotencyKey string,
	createdAt time.Time,
) CheckInCompletedEvent {
	return CheckInCompletedEvent{
		UserID:         userID,
		HabitID:        habitID,
		CheckedDate:    checkedDate,
		PointAmount:    pointAmount,
		PointReason:    pointReason,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      createdAt,
	}
}

func (e CheckInCompletedEvent) EventType() string {
	return EventTypeCheckInCompleted
}

func (e CheckInCompletedEvent) OccurredAt() time.Time {
	return e.CreatedAt
}
