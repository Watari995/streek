package cache

import (
	"context"

	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type IStreakCache interface {
	Get(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) (*StreakSnapshot, bool, error)
	Set(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString, snapshot StreakSnapshot) error
	Invalidate(ctx context.Context, habitID valueobject.HabitID, date valueobject.DateString) error
}

type StreakSnapshot struct {
	CurrentStreak int
	LongestStreak int
	CheckedToday  bool
}
