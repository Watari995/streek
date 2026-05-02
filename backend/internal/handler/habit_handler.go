package handler

import (
	"net/http"
	"time"

	"github.com/Watari995/streek/backend/internal/apperror"
	applicationHabit "github.com/Watari995/streek/backend/internal/application/habit"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/middleware"
	"github.com/labstack/echo/v4"
)

type HabitHandler struct {
	list   *applicationHabit.List
	create *applicationHabit.Create
	update *applicationHabit.Update
	delete *applicationHabit.Delete
}

// Request DTOs
type createRequest struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LabelColor  string `json:"label_color"`
}

type updateRequest struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LabelColor  string `json:"label_color"`
}

type deleteRequest struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

// Response DTOs
type habitResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description *string    `json:"description"`
	LabelColor  string    `json:"label_color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toHabitResponse(habit entity.Habit) habitResponse {
	return habitResponse{
		ID:          habit.ID().String(),
		UserID:      habit.UserID().String(),
		Name:        habit.Name().String(),
		Description: habit.Description().String(),
		LabelColor:  habit.LabelColor().String(),
		CreatedAt:   habit.CreatedAt(),
		UpdatedAt:   habit.UpdatedAt(),
	}
}

func NewHabitHandler(list *applicationHabit.List, create *applicationHabit.Create, update *applicationHabit.Update, delete *applicationHabit.Delete) *HabitHandler {
	return &HabitHandler{
		list:   list,
		create: create,
		update: update,
		delete: delete,
	}
}

func (h *HabitHandler) List(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}

	// call application service
	output, err := h.list.Do(ctx, applicationHabit.ListInput{
		UserID: userID,
	})
	if err != nil {
		return RespondError(c, err)
	}

	// return response
	var habits []habitResponse
	for _, habit := range output {
		habits = append(habits, toHabitResponse(*habit))
	}
	return c.JSON(http.StatusOK, habits)
}
