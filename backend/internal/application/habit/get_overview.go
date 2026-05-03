package habit

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/repository"
	"github.com/Watari995/streek/backend/internal/domain/service"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

type GetOverview struct {
	habitRepo     repository.IHabitRepository
	checkInRepo   repository.ICheckInRepository
	streakService *service.StreakService
}

type GetOverviewInput struct {
	UserID valueobject.UserID
	Today  valueobject.DateString
}

type HabitOverview struct {
	Habit         *entity.Habit
	CurrentStreak int
	LongestStreak int
	CheckedToday  bool
}

type GetOverviewOutput struct {
	Habits        []HabitOverview
	LongestStreak int // longest streak across all habits
	DoneToday     int
}

func NewGetOverview(habitRepo repository.IHabitRepository, checkInRepo repository.ICheckInRepository, streakService *service.StreakService) *GetOverview {
	return &GetOverview{
		habitRepo:     habitRepo,
		checkInRepo:   checkInRepo,
		streakService: streakService,
	}
}

func (s *GetOverview) Do(ctx context.Context, input GetOverviewInput) (GetOverviewOutput, error) {
	habits, err := s.habitRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return GetOverviewOutput{}, errors.Wrap(err, "failed to find habits")
	}

	if len(habits) == 0 {
		return GetOverviewOutput{Habits: []HabitOverview{}}, nil
	}

	// create errgroup to wait for all goroutines to complete
	g, gctx := errgroup.WithContext(ctx)
	results := make([]HabitOverview, len(habits))

	for i, habit := range habits {
		// capture loop variables
		i, h := i, habit
		// start goroutine for each habit
		g.Go(func() error {
			checkIns, err := s.checkInRepo.FindByHabitID(gctx, h.ID())
			if err != nil {
				return errors.Wrap(err, "failed to find check-ins")
			}
			// compute overview for habit
			results[i] = HabitOverview{
				Habit:         h,
				CurrentStreak: s.streakService.ComputeCurrentStreak(checkIns, input.Today),
				LongestStreak: s.streakService.ComputeLongestStreak(checkIns),
				CheckedToday:  s.streakService.HasCheckInOnDate(checkIns, input.Today),
			}
			return nil
		})
	}

	// wait til all goroutines complete or error occurs
	if err := g.Wait(); err != nil {
		return GetOverviewOutput{}, errors.Wrap(err, "failed to compute habits overview")
	}

	longest := 0
	doneToday := 0
	for _, result := range results {
		if result.LongestStreak > longest {
			longest = result.LongestStreak
		}
		if result.CheckedToday {
			doneToday++
		}
	}

	return GetOverviewOutput{
		Habits:        results,
		LongestStreak: longest,
		DoneToday:     doneToday,
	}, nil
}
