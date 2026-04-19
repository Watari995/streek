package habit

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type Create struct {
	habitRepo repository.IHabitRepository
}

type CreateInput struct {
	UserID      valueobject.UserID
	Name        valueobject.String50
	Description *valueobject.String200
	LabelColor  valueobject.HexColor
}

func (h *Create) Do(
	ctx context.Context,
	input CreateInput,
) (*entity.Habit, error) {
	habitEntity := entity.CreateHabit(
		input.UserID,
		input.Name,
		input.Description,
		input.LabelColor,
	)
	habit, err := h.habitRepo.Save(
		ctx,
		habitEntity,
	)
	if err != nil {
		return nil, errors.Wrap(err, "internal server error")
	}
	return habit, nil
}

func NewCreate(habitRepo repository.IHabitRepository) *Create {
	return &Create{habitRepo: habitRepo}
}
