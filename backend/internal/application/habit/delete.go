package habit

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type Delete struct {
	habitRepo repository.IHabitRepository
}

type DeleteInput struct {
	ID     valueobject.HabitID
	UserID valueobject.UserID
}

func (h *Delete) Do(ctx context.Context, input DeleteInput) error {
	habit, err := h.habitRepo.FindByID(ctx, input.ID)
	if err != nil {
		return errors.Wrap(err, "failed to find habit")
	}
	if habit.UserID() != input.UserID {
		return errors.New("forbidden access")
	}
	err = h.habitRepo.Delete(ctx, input.ID)
	if err != nil {
		return errors.Wrap(err, "failed to delete habit")
	}
	return nil
}

func NewDelete(habitRepo repository.IHabitRepository) *Delete {
	return &Delete{habitRepo: habitRepo}
}
