package handler

import (
	"net/http"
	"time"

	"github.com/Watari995/streek/backend/internal/apperror"
	applicationHabit "github.com/Watari995/streek/backend/internal/application/habit"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/Watari995/streek/backend/internal/middleware"
	"github.com/labstack/echo/v4"
)

type StatsHandler struct {
	getOverview *applicationHabit.GetOverview
}

// Response DTOs
type habitOverviewResponse struct {
	HabitID       string `json:"habit_id"`
	HabitName     string `json:"habit_name"`
	LabelColor    string `json:"label_color"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
	CheckedToday  bool   `json:"checked_today"`
}

type statsOverviewResponse struct {
	Habits        []habitOverviewResponse `json:"habits"`
	LongestStreak int                     `json:"longest_streak"`
	DoneToday     int                     `json:"done_today"`
}

func toHabitOverviewResponse(habit applicationHabit.HabitOverview) habitOverviewResponse {
	return habitOverviewResponse{
		HabitID:       habit.Habit.ID().String(),
		HabitName:     habit.Habit.Name().String(),
		LabelColor:    habit.Habit.LabelColor().String(),
		CurrentStreak: habit.CurrentStreak,
		LongestStreak: habit.LongestStreak,
		CheckedToday:  habit.CheckedToday,
	}
}

func NewStatsHandler(getOverview *applicationHabit.GetOverview) *StatsHandler {
	return &StatsHandler{
		getOverview: getOverview,
	}
}

func (h *StatsHandler) GetOverview(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}

	todayStr := c.QueryParam("today")
	if todayStr == "" {
		todayStr = time.Now().Format("2006-01-02")
	}
	today, err := valueobject.NewDateStringFromString(todayStr)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid today"))
	}

	// call application service
	output, err := h.getOverview.Do(ctx, applicationHabit.GetOverviewInput{
		UserID: userID,
		Today:  today,
	})
	if err != nil {
		return RespondError(c, err)
	}

	// return response
	habits := make([]habitOverviewResponse, len(output.Habits))
	for i, habit := range output.Habits {
		habits[i] = toHabitOverviewResponse(habit)
	}
	return c.JSON(http.StatusOK, statsOverviewResponse{
		Habits:        habits,
		LongestStreak: output.LongestStreak,
		DoneToday:     output.DoneToday,
	})
}
