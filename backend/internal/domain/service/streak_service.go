package service

import (
	"sort"
	"time"

	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
)

type StreakService struct{}

func NewStreakService() *StreakService {
	return &StreakService{}
}

func (s *StreakService) ComputeCurrentStreak(checkIns []*entity.CheckIn, today valueobject.DateString) int {
	if len(checkIns) == 0 {
		return 0
	}

	daySet := make(map[string]bool)
	for _, checkIn := range checkIns {
		daySet[checkIn.CheckedDate().String()] = true
	}

	todayTime, err := time.Parse("2006-01-02", today.String())
	if err != nil {
		return 0
	}
	// start first cursor
	var cursor time.Time
	if daySet[todayTime.Format("2006-01-02")] {
		cursor = todayTime
	} else {
		yesterday := todayTime.AddDate(0, 0, -1)
		if daySet[yesterday.Format("2006-01-02")] {
			cursor = yesterday
		} else {
			return 0
		}
	}

	// compute from cursor in reverse order
	count := 0
	for daySet[cursor.Format("2006-01-02")] {
		count++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return count
}

func (s *StreakService) ComputeLongestStreak(checkIns []*entity.CheckIn) int {
	if len(checkIns) == 0 {
		return 0
	}

	// sort checkIns by checked date
	dates := make([]time.Time, 0, len(checkIns))
	for _, checkIn := range checkIns {
		date, err := time.Parse("2006-01-02", checkIn.CheckedDate().String())
		if err != nil {
			continue
		}
		dates = append(dates, date)
	}
	// sort dates in ascending order
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	// compute longest streak
	longest := 1
	current := 1
	for i := 1; i < len(dates); i++ {
		if dates[i].Equal(dates[i-1].AddDate(0, 0, 1)) {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}
	return longest
}

func (s *StreakService) HasCheckInOnDate(checkIns []*entity.CheckIn, date valueobject.DateString) bool {
	for _, checkIn := range checkIns {
		if checkIn.CheckedDate() == date {
			return true
		}
	}
	return false
}
