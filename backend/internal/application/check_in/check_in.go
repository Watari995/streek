package checkin

import (
	"context"
	"time"

	"github.com/Watari995/streek/backend/internal/apperror"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type CheckIn struct {
	checkInRepo repository.ICheckInRepository
	habitRepo   repository.IHabitRepository
}

type CheckInInput struct {
	HabitID valueobject.HabitID
	UserID  valueobject.UserID
}

func (c *CheckIn) Do(ctx context.Context, input CheckInInput) error {
	habit, err := c.habitRepo.FindByID(ctx, input.HabitID)
	if err != nil {
		return apperror.NewInternalServerError().SetMessage("failed to find habit")
	}
	if habit == nil {
		return apperror.NewNotFoundError().SetMessage("habit not found")
	}
	if habit.UserID() != input.UserID {
		return apperror.NewForbiddenError().SetMessage("you do not have permission to check in this habit")
	}

	// create check in entity and save it
	checkInEntity := entity.CreateCheckIn(
		input.HabitID,
		// now should be  JST, so must check
		time.Now(),
	)
	if _, err := c.checkInRepo.Save(ctx, checkInEntity); err != nil {
		return apperror.NewInternalServerError().SetMessage("failed to check in")
	}
	return nil
}

func NewCheckIn(checkInRepo repository.ICheckInRepository) *CheckIn {
	return &CheckIn{checkInRepo: checkInRepo}
}
