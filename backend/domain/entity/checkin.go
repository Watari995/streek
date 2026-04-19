package entity

import (
	"time"

	"github.com/Watari995/streek/backend/domain/valueobject"
)

type CheckIn struct {
	id          valueobject.CheckInID
	habitID     valueobject.HabitID
	checkedDate time.Time
	createdAt   time.Time
}

func (c CheckIn) ID() valueobject.CheckInID {
	return c.id
}

func (c CheckIn) HabitID() valueobject.HabitID {
	return c.habitID
}

func (c CheckIn) CheckedDate() time.Time {
	return c.checkedDate
}

func (c CheckIn) CreatedAt() time.Time {
	return c.createdAt
}

// DB restoration
func NewCheckIn(
	id valueobject.CheckInID,
	habitID valueobject.HabitID,
	checkedDate time.Time,
	createdAt time.Time,
) CheckIn {
	return CheckIn{
		id:          id,
		habitID:     habitID,
		checkedDate: checkedDate,
		createdAt:   createdAt,
	}
}

// New creation
func CreateCheckIn(
	habitID valueobject.HabitID,
	checkedDate time.Time,
) CheckIn {
	return NewCheckIn(
		valueobject.NewCheckInID(),
		habitID,
		checkedDate,
		time.Now(),
	)
}
