package types

import (
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

const (
	EventTypeCheckInSucceeded = "checkInSucceeded"
)

type CheckInSucceededEvent struct {
	UserID      valueobject.UserID
	HabitID     valueobject.HabitID
	CheckedDate valueobject.DateString
	CreatedAt   time.Time
}

func NewCheckInSucceededEvent(
	userID valueobject.UserID,
	habitID valueobject.HabitID,
	checkedDate valueobject.DateString,
	createdAt time.Time,
) CheckInSucceededEvent {
	return CheckInSucceededEvent{
		UserID:      userID,
		HabitID:     habitID,
		CheckedDate: checkedDate,
		CreatedAt:   createdAt,
	}
}

func (e CheckInSucceededEvent) EventType() string {
	return EventTypeCheckInSucceeded
}

func (e CheckInSucceededEvent) OccurredAt() time.Time {
	return e.CreatedAt
}
