package handler

import (
	"net/http"
	"time"

	"github.com/Watari995/streek/backend/internal/apperror"
	applicationHabit "github.com/Watari995/streek/backend/internal/application/habit"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
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
	Name        string  `json:"name"`
	Description *string `json:"description"`
	LabelColor  string  `json:"label_color"`
}

type updateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	LabelColor  string  `json:"label_color"`
}

// Response DTOs
type habitResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	LabelColor  string    `json:"label_color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// prepared for pagination
type habitListResponse struct {
	Habits []habitResponse `json:"habits"`
}

func toHabitResponse(habit entity.Habit) habitResponse {
	var description *string
	if d := habit.Description(); d != nil {
		s := d.String()
		description = &s
	}
	return habitResponse{
		ID:          habit.ID().String(),
		UserID:      habit.UserID().String(),
		Name:        habit.Name().String(),
		Description: description,
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
	return c.JSON(http.StatusOK, habitListResponse{Habits: habits})
}

func (h *HabitHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}

	// bind request body
	var req createRequest
	if err := c.Bind(&req); err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid request body"))
	}

	// value object conversion
	name, err := valueobject.NewString50(req.Name)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid name"))
	}
	var description *valueobject.String200
	if req.Description != nil {
		description, err = valueobject.NewString200(*req.Description)
		if err != nil {
			return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid description"))
		}
	}
	labelColor, err := valueobject.NewHexColor(req.LabelColor)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid label color"))
	}

	// call application service
	output, err := h.create.Do(ctx, applicationHabit.CreateInput{
		UserID:      userID,
		Name:        name,
		Description: description,
		LabelColor:  labelColor,
	})
	if err != nil {
		return RespondError(c, err)
	}

	// return response
	return c.JSON(http.StatusCreated, toHabitResponse(*output))
}

func (h *HabitHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}
	habitID, err := valueobject.NewHabitIDFromString(c.Param("id"))
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid habit ID"))
	}

	// bind request body
	var req updateRequest
	if err := c.Bind(&req); err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid request body"))
	}

	// value object conversion
	name, err := valueobject.NewString50(req.Name)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid name"))
	}
	var description *valueobject.String200
	if req.Description != nil {
		description, err = valueobject.NewString200(*req.Description)
		if err != nil {
			return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid description"))
		}
	}
	labelColor, err := valueobject.NewHexColor(req.LabelColor)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid label color"))
	}

	// call application service
	output, err := h.update.Do(ctx, applicationHabit.UpdateInput{
		ID:          habitID,
		UserID:      userID,
		Name:        name,
		Description: description,
		LabelColor:  labelColor,
	})
	if err != nil {
		return RespondError(c, err)
	}

	// return response
	return c.JSON(http.StatusOK, toHabitResponse(*output))
}

func (h *HabitHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}
	habitID, err := valueobject.NewHabitIDFromString(c.Param("id"))
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid habit ID"))
	}

	// call application service
	err = h.delete.Do(ctx, applicationHabit.DeleteInput{
		ID:     habitID,
		UserID: userID,
	})
	if err != nil {
		return RespondError(c, err)
	}

	// return response (must not send a message body when no content so use c.NoContent())
	return c.NoContent(http.StatusNoContent)
}
