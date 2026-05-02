package habit

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type List struct {
	habitRepo repository.IHabitRepository
}

type ListInput struct {
	UserID valueobject.UserID
}

func (l *List) Do(ctx context.Context, input ListInput) ([]*entity.Habit, error) {
	habits, err := l.habitRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find habits")
	}
	return habits, nil
}

func NewList(habitRepo repository.IHabitRepository) *List {
	return &List{habitRepo: habitRepo}
}
