package checkin

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/apperror"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type Undo struct {
	checkInRepo repository.ICheckInRepository
	habitRepo   repository.IHabitRepository
}

type UndoInput struct {
	HabitID valueobject.HabitID
	UserID  valueobject.UserID
}

func (u *Undo) Do(ctx context.Context, input UndoInput) error {
	habit, err := u.habitRepo.FindByID(ctx, input.HabitID)
	if err != nil {
		return apperror.NewInternalServerError().SetMessage("failed to find habit")
	}
	if habit.UserID() != input.UserID {
		return apperror.NewForbiddenError().SetMessage("you do not have permission to undo check in this habit")
	}
	// delete check in
	err = u.checkInRepo.DeleteByHabitIDAndCheckedDate(ctx, input.HabitID, time.Now())
	if err != nil {
		return apperror.NewInternalServerError().SetMessage("failed to undo check in")
	}
	return nil
}

func NewUndo(checkInRepo repository.ICheckInRepository, habitRepo repository.IHabitRepository) *Undo {
	return &Undo{checkInRepo: checkInRepo, habitRepo: habitRepo}
}
