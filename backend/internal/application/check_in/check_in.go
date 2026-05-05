package checkin

import (
	"context"

	"github.com/Watari995/streek/backend/internal/apperror"
	domainCache "github.com/Watari995/streek/backend/internal/domain/cache"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/transaction"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type CheckIn struct {
	checkInRepo     repository.ICheckInRepository
	habitRepo       repository.IHabitRepository
	streakCache     domainCache.IStreakCache
	pointLedgerRepo repository.IPointLedgerRepository
	txManager       transaction.ITransactionManager
}

type CheckInInput struct {
	HabitID     valueobject.HabitID
	UserID      valueobject.UserID
	CheckedDate valueobject.DateString
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
		input.CheckedDate,
	)

	err = c.txManager.Run(ctx, func(ctx context.Context) error {
		if _, err := c.checkInRepo.Save(ctx, checkInEntity); err != nil {
			return apperror.NewInternalServerError().SetMessage("failed to check in")
		}
		pointLedgerEntity := entity.CreatePointLedger(
			input.UserID,
			&input.HabitID,
			valueobject.NewPointTypeEarn(),
			checkInEntity.PointAmount(),
			checkInEntity.PointReason(),
			checkInEntity.IdempotencyKey(),
		)
		if _, err := c.pointLedgerRepo.Save(ctx, pointLedgerEntity); err != nil {
			return apperror.NewInternalServerError().SetMessage("failed to save point ledger")
		}
		return nil
	})
	if err != nil {
		return err
	}

	// clear streak cache
	_ = c.streakCache.Invalidate(ctx, input.HabitID, input.CheckedDate)

	return nil
}

func NewCheckIn(checkInRepo repository.ICheckInRepository, habitRepo repository.IHabitRepository, streakCache domainCache.IStreakCache, pointLedgerRepo repository.IPointLedgerRepository, txManager transaction.ITransactionManager) *CheckIn {
	return &CheckIn{checkInRepo: checkInRepo, habitRepo: habitRepo, streakCache: streakCache, pointLedgerRepo: pointLedgerRepo, txManager: txManager}
}
