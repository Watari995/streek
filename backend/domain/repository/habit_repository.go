package repository

import (
	"context"

	"github.com/Watari995/streek/backend/domain/entity"
	"github.com/Watari995/streek/backend/domain/valueobject"
)

type IHabitRepository interface {
	Save(context.Context, entity.Habit) (*entity.Habit, error)
	FindByID(context.Context, valueobject.HabitID) (*entity.Habit, error)
	FindByUserID(context.Context, valueobject.UserID) ([]*entity.Habit, error)
	Delete(context.Context, valueobject.HabitID) error
}
