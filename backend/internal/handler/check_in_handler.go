package handler

import (
	"net/http"

	"github.com/Watari995/streek/backend/internal/apperror"
	applicationCheckIn "github.com/Watari995/streek/backend/internal/application/check_in"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/Watari995/streek/backend/internal/middleware"
	"github.com/labstack/echo/v4"
)

type CheckInHandler struct {
	checkIn *applicationCheckIn.CheckIn
	undo    *applicationCheckIn.Undo
}

// Request DTOs
type checkInRequest struct {
	CheckedDate string `json:"checked_date"`
}

func NewCheckInHandler(checkIn *applicationCheckIn.CheckIn, undo *applicationCheckIn.Undo) *CheckInHandler {
	return &CheckInHandler{checkIn: checkIn, undo: undo}
}

// helper
func parseCheckInRequest(c echo.Context) (valueobject.HabitID, valueobject.DateString, error) {
	habitID, err := valueobject.NewHabitIDFromString(c.Param("id"))
	if err != nil {
		return valueobject.HabitID{}, valueobject.DateString{}, apperror.NewBadRequestError().SetMessage("invalid habit ID")
	}

	var req checkInRequest
	if err := c.Bind(&req); err != nil {
		return valueobject.HabitID{}, valueobject.DateString{}, apperror.NewBadRequestError().SetMessage("invalid request body")
	}

	checkedDate, err := valueobject.NewDateStringFromString(req.CheckedDate)
	if err != nil {
		return valueobject.HabitID{}, valueobject.DateString{}, apperror.NewBadRequestError().SetMessage("invalid checked date")
	}

	return habitID, checkedDate, nil
}

func (h *CheckInHandler) CheckIn(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}

	habitID, checkedDate, err := parseCheckInRequest(c)
	if err != nil {
		return RespondError(c, err)
	}
	err = h.checkIn.Do(ctx, applicationCheckIn.CheckInInput{
		HabitID:     habitID,
		UserID:      userID,
		CheckedDate: checkedDate,
	})

	if err != nil {
		return RespondError(c, err)
	}

	// return response
	return c.NoContent(http.StatusNoContent)
}

func (h *CheckInHandler) Undo(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}

	habitID, checkedDate, err := parseCheckInRequest(c)
	if err != nil {
		return RespondError(c, err)
	}
	err = h.undo.Do(ctx, applicationCheckIn.UndoInput{
		HabitID:     habitID,
		UserID:      userID,
		CheckedDate: checkedDate,
	})

	if err != nil {
		return RespondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
