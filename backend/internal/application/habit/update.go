package habit

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type Update struct {
	habitRepo repository.IHabitRepository
}

type UpdateInput struct {
	ID          valueobject.HabitID
	UserID      valueobject.UserID
	Name        valueobject.String50
	Description *valueobject.String200
	LabelColor  valueobject.HexColor
}

func (h *Update) Do(
	ctx context.Context,
	input UpdateInput,
) (*entity.Habit, error) {
	habit, err := h.habitRepo.FindByID(
		ctx,
		input.ID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find habit")
	}
	habit.SetName(input.Name)
	habit.SetDescription(input.Description)
	habit.SetLabelColor(input.LabelColor)

	updatedHabit, err := h.habitRepo.Save(
		ctx,
		*habit,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update habit")
	}
	return updatedHabit, nil
}

func NewUpdate(habitRepo repository.IHabitRepository) *Update {
	return &Update{habitRepo: habitRepo}
}
