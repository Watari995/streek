package eventhandler

import (
	"context"
	"fmt"

	"github.com/Watari995/streek/backend/internal/apperror"
	"github.com/Watari995/streek/backend/internal/domain/event"
	"github.com/Watari995/streek/backend/internal/domain/event/types"
	domainnotification "github.com/Watari995/streek/backend/internal/domain/notification"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/service"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
)

type NotifyStreakMilestone struct {
	notifier      domainnotification.INotifier
	checkInRepo   repository.ICheckInRepository
	streakService *service.StreakService
	notifyTo      valueobject.Email
}

func NewNotifyStreakMilestone(notifier domainnotification.INotifier, checkInRepo repository.ICheckInRepository, streakService *service.StreakService, notifyTo valueobject.Email) *NotifyStreakMilestone {
	return &NotifyStreakMilestone{notifier: notifier, checkInRepo: checkInRepo, streakService: streakService, notifyTo: notifyTo}
}

func (h *NotifyStreakMilestone) Handle(ctx context.Context, event event.DomainEvent) error {
	checkInSucceededEvent, ok := event.(types.CheckInSucceededEvent)
	if !ok {
		return apperror.NewInternalServerError().SetMessage("invalid event type")
	}

	checkIns, err := h.checkInRepo.FindByHabitID(ctx, checkInSucceededEvent.HabitID)
	if err != nil {
		return errors.Wrap(err, "failed to find check-ins")
	}

	currentStreak := h.streakService.ComputeCurrentStreak(checkIns, checkInSucceededEvent.CheckedDate)
	if !h.streakService.IsStreakMilestone(currentStreak) {
		return nil
	}

	subject := "🎉 You've reached a streak milestone! 🎉"
	body := fmt.Sprintf("You've achieved a streak of %d days! Keep it up!", currentStreak)
	return h.notifier.Notify(ctx, h.notifyTo.String(), subject, body)
}
