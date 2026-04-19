package entity

import (
	"time"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type Habit struct {
	id          valueobject.HabitID
	userID      valueobject.UserID
	name        valueobject.String50
	description *valueobject.String200
	labelColor  valueobject.HexColor
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

func (h *Habit) SetName(name valueobject.String50) {
	h.name = name
}

func (h Habit) Description() *valueobject.String200 {
	return h.description
}

func (h *Habit) SetDescription(description *valueobject.String200) {
	h.description = description
}

func (h Habit) LabelColor() valueobject.HexColor {
	return h.labelColor
}

func (h *Habit) SetLabelColor(labelColor valueobject.HexColor) {
	h.labelColor = labelColor
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
	labelColor valueobject.HexColor,
	createdAt time.Time,
	updatedAt time.Time,
) Habit {
	return Habit{
		id:          id,
		userID:      userID,
		name:        name,
		description: description,
		labelColor:  labelColor,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func CreateHabit(
	userID valueobject.UserID,
	name valueobject.String50,
	description *valueobject.String200,
	labelColor valueobject.HexColor,
) Habit {
	return NewHabit(
		valueobject.NewHabitID(),
		userID,
		name,
		description,
		labelColor,
		time.Now(),
		time.Now(),
	)
}
