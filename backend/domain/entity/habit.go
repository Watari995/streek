package entity

import "github.com/Watari995/streek/backend/domain/valueobject"

type Habit struct {
	id valueobject.HabitID
	userId valueobject.UserID
	name val
}