package service_test

import (
	"testing"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/service"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCheckIns(t *testing.T, dates []string) []*entity.CheckIn {
	t.Helper()
	cs := make([]*entity.CheckIn, 0, len(dates))
	for _, d := range dates {
		ds, err := valueobject.NewDateStringFromString(d)
		require.NoError(t, err)
		c := entity.NewCheckIn(
			valueobject.NewCheckInID(),
			valueobject.NewHabitID(),
			ds,
			time.Now(),
		)
		cs = append(cs, &c)
	}
	return cs
}

func TestStreakService_ComputeCurrentStreak(t *testing.T) {
	t.Parallel()
	s := service.NewStreakService()

	cases := []struct {
		name  string
		dates []string
		today string
		want  int
	}{
		{"empty", nil, "2026-05-03", 0},
		{"today only", []string{"2026-05-03"}, "2026-05-03", 1},
		{"3 consecutive ending today", []string{"2026-05-01", "2026-05-02", "2026-05-03"}, "2026-05-03", 3},
		{"yesterday only (grace period)", []string{"2026-05-02"}, "2026-05-03", 1},
		{"2-day gap (no streak)", []string{"2026-05-01"}, "2026-05-03", 0},
		{"break in middle - count from latest run", []string{"2026-04-01", "2026-04-02", "2026-05-02", "2026-05-03"}, "2026-05-03", 2},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			today, err := valueobject.NewDateStringFromString(tc.today)
			require.NoError(t, err)
			checkIns := makeCheckIns(t, tc.dates)
			got := s.ComputeCurrentStreak(checkIns, today)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestStreakService_ComputeLongestStreak(t *testing.T) {
	t.Parallel()
	s := service.NewStreakService()

	cases := []struct {
		name  string
		dates []string
		want  int
	}{
		{"empty", nil, 0},
		{"single", []string{"2026-05-01"}, 1},
		{"3 consecutive", []string{"2026-05-01", "2026-05-02", "2026-05-03"}, 3},
		{"two runs - take longer", []string{"2026-05-01", "2026-05-02", "2026-05-04", "2026-05-05", "2026-05-06"}, 3},
		{"unsorted input still works", []string{"2026-05-03", "2026-05-01", "2026-05-02"}, 3},
		{"long history with gaps", []string{"2026-04-01", "2026-04-05", "2026-04-06", "2026-04-07", "2026-05-01"}, 3},
		{"all isolated days", []string{"2026-04-01", "2026-04-03", "2026-04-05"}, 1},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkIns := makeCheckIns(t, tc.dates)
			got := s.ComputeLongestStreak(checkIns)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestStreakService_HasCheckInOnDate(t *testing.T) {
	t.Parallel()
	s := service.NewStreakService()

	cases := []struct {
		name  string
		dates []string
		date  string
		want  bool
	}{
		{"empty check-ins", nil, "2026-05-01", false},
		{"date matches the only one", []string{"2026-05-01"}, "2026-05-01", true},
		{"date matches one of many", []string{"2026-05-01", "2026-05-02", "2026-05-03"}, "2026-05-02", true},
		{"date matches none", []string{"2026-05-01", "2026-05-03"}, "2026-05-02", false},
		{"date later than all check-ins", []string{"2026-05-01", "2026-05-02"}, "2026-05-10", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			date, err := valueobject.NewDateStringFromString(tc.date)
			require.NoError(t, err)
			checkIns := makeCheckIns(t, tc.dates)
			got := s.HasCheckInOnDate(checkIns, date)
			assert.Equal(t, tc.want, got)
		})
	}
}
