package repository

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type ICheckInRepository interface {
	Save(context.Context, entity.CheckIn) (*entity.CheckIn, error)
	FindByHabitID(context.Context, valueobject.HabitID) ([]*entity.CheckIn, error)
	Delete(context.Context, valueobject.CheckInID) error
}
