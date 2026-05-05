package eventhandler

import (
	"context"

	"github.com/Watari995/streek/backend/internal/apperror"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/event"
	"github.com/Watari995/streek/backend/internal/domain/event/types"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type EarnPointsOnCheckIn struct {
	pointLedgerRepo repository.IPointLedgerRepository
}

func NewEarnPointsOnCheckIn(pointLedgerRepo repository.IPointLedgerRepository) *EarnPointsOnCheckIn {
	return &EarnPointsOnCheckIn{pointLedgerRepo: pointLedgerRepo}
}

func (h *EarnPointsOnCheckIn) Handle(ctx context.Context, event event.DomainEvent) error {
	checkInCompletedEvent, ok := event.(*types.CheckInCompletedEvent)
	if !ok {
		return apperror.NewInternalServerError().SetMessage("invalid event type")
	}

	pointLedgerEntity := entity.CreatePointLedger(
		checkInCompletedEvent.UserID,
		&checkInCompletedEvent.HabitID,
		valueobject.NewPointTypeEarn(),
		checkInCompletedEvent.PointAmount,
		checkInCompletedEvent.PointReason,
		checkInCompletedEvent.IdempotencyKey,
	)
	if _, err := h.pointLedgerRepo.Save(ctx, pointLedgerEntity); err != nil {
		return apperror.NewInternalServerError().SetMessage("failed to save point ledger")
	}
	return nil
}
