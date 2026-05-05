package entity

import (
	"fmt"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

const (
	pointAmountPerCheckIn     = 10
	pointReasonCheckIn        = "checkIn"
	pointIdempotencyKeyPrefix = "checkIn"
)

type CheckIn struct {
	id          valueobject.CheckInID
	habitID     valueobject.HabitID
	checkedDate valueobject.DateString
	createdAt   time.Time
}

func (c *CheckIn) ID() valueobject.CheckInID {
	return c.id
}

func (c *CheckIn) HabitID() valueobject.HabitID {
	return c.habitID
}

func (c *CheckIn) CheckedDate() valueobject.DateString {
	return c.checkedDate
}

func (c *CheckIn) CreatedAt() time.Time {
	return c.createdAt
}

// point ledger related
func (c *CheckIn) PointAmount() valueobject.PositiveInt {
	return valueobject.MustPositiveInt(pointAmountPerCheckIn)
}

func (c *CheckIn) PointReason() valueobject.String50 {
	return valueobject.MustString50(pointReasonCheckIn)
}

func (c *CheckIn) IdempotencyKey() string {
	return fmt.Sprintf("%s:%s:%s", pointIdempotencyKeyPrefix, c.habitID.String(), c.checkedDate.String())
}

// DB restoration
func NewCheckIn(
	id valueobject.CheckInID,
	habitID valueobject.HabitID,
	checkedDate valueobject.DateString,
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
	checkedDate valueobject.DateString,
) CheckIn {
	return NewCheckIn(
		valueobject.NewCheckInID(),
		habitID,
		checkedDate,
		time.Now(),
	)
}
