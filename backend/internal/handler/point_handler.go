package handler

import (
	"net/http"
	"time"

	"github.com/Watari995/streek/backend/internal/apperror"
	applicationPoint "github.com/Watari995/streek/backend/internal/application/point"
	"github.com/Watari995/streek/backend/internal/middleware"
	"github.com/labstack/echo/v4"
)

type PointHandler struct {
	getBalance *applicationPoint.GetBalance
	getHistory *applicationPoint.GetHistory
}

// Response DTOs

type pointBalanceResponse struct {
	Balance int `json:"balance"`
}

type pointHistoryEntryResponse struct {
	ID             string    `json:"id"`
	HabitID        *string   `json:"habit_id"`
	Type           string    `json:"type"`
	Amount         int       `json:"amount"`
	Reason         string    `json:"reason"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type pointHistoryResponse struct {
	Entries []pointHistoryEntryResponse `json:"entries"`
}

func NewPointHandler(
	getBalance *applicationPoint.GetBalance,
	getHistory *applicationPoint.GetHistory,
) *PointHandler {
	return &PointHandler{
		getBalance: getBalance,
		getHistory: getHistory,
	}
}

func (h *PointHandler) GetBalance(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}

	output, err := h.getBalance.Do(ctx, applicationPoint.GetBalanceInput{UserID: userID})
	if err != nil {
		return RespondError(c, err)
	}
	return c.JSON(http.StatusOK, pointBalanceResponse{Balance: output.Balance})
}

func (h *PointHandler) GetHistory(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return RespondError(c, apperror.NewUnauthorizedError().SetMessage("unauthorized"))
	}

	output, err := h.getHistory.Do(ctx, applicationPoint.GetHistoryInput{UserID: userID})
	if err != nil {
		return RespondError(c, err)
	}

	entries := make([]pointHistoryEntryResponse, len(output.Entries))
	for i, e := range output.Entries {
		var habitID *string
		if hid := e.HabitID(); hid != nil {
			s := hid.String()
			habitID = &s
		}
		entries[i] = pointHistoryEntryResponse{
			ID:             e.ID().String(),
			HabitID:        habitID,
			Type:           e.PointType().String(),
			Amount:         e.Amount().Int(),
			Reason:         e.Reason().String(),
			IdempotencyKey: e.IdempotencyKey(),
			CreatedAt:      e.CreatedAt(),
		}
	}
	return c.JSON(http.StatusOK, pointHistoryResponse{Entries: entries})
}
