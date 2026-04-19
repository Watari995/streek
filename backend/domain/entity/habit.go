package entity

import (
	"time"

	"github.com/Watari995/streek/backend/domain/valueobject"
)

type Habit struct {
	id          valueobject.HabitID
	userID      valueobject.UserID
	name        valueobject.String50
	description *valueobject.String200
	color       valueobject.HexColor
	createdAt   time.Time
	updatedAt   time.Time
}

func (h Habit) ID() valueobject.HabitID {
	return h.id
}

func (h Habit) UserID() valueobject.UserID {
	return h.userID
}

func (h Habit) Name() valueobject.String50 {
	return h.name
}

func (h Habit) Description() *valueobject.String200 {
	return h.description
}

func (h Habit) Color() valueobject.HexColor {
	return h.color
}

func (h Habit) CreatedAt() time.Time {
	return h.createdAt
}

func (h Habit) UpdatedAt() time.Time {
	return h.updatedAt
}

func NewHabit(
	id valueobject.HabitID,
	userID valueobject.UserID,
	name valueobject.String50,
	description *valueobject.String200,
	color valueobject.HexColor,
	createdAt time.Time,
	updatedAt time.Time,
) Habit {
	return Habit{
		id:          id,
		userID:      userID,
		name:        name,
		description: description,
		color:       color,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func CreateHabit(
	userID valueobject.UserID,
	name valueobject.String50,
	description *valueobject.String200,
	color valueobject.HexColor,
) Habit {
	return NewHabit(
		valueobject.NewHabitID(),
		userID,
		name,
		description,
		color,
		time.Now(),
		time.Now(),
	)
}
