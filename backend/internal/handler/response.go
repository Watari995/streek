package handler

import (
	"net/http"

	"github.com/Watari995/streek/backend/internal/apperror"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RespondError(c echo.Context, err error) error {
	myErr, ok := apperror.AsMyError(err)
	if !ok {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorBody{
				Code:    string(apperror.CodeInternalError),
				Message: "An unexpected error occurred",
			},
		})
	}
	return c.JSON(myErr.Status(), ErrorResponse{
		Error: ErrorBody{
			Code:    string(myErr.Code()),
			Message: myErr.Message(),
		},
	})
}
