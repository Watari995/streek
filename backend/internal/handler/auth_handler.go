package handler

import (
	"net/http"
	"time"

	"github.com/Watari995/streek/backend/internal/apperror"
	applicationAuth "github.com/Watari995/streek/backend/internal/application/auth"
	"github.com/Watari995/streek/backend/internal/domain/entity"
	"github.com/Watari995/streek/backend/internal/domain/valueobject"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	register *applicationAuth.Register
	login    *applicationAuth.Login
}

// Request DTOs (unexported)
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response DTOs (unexported)
type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type authResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

func toUserResponse(user entity.User) userResponse {
	return userResponse{
		ID:        user.ID().String(),
		Email:     user.Email().String(),
		CreatedAt: user.CreatedAt(),
	}
}

// handler method
func NewAuthHandler(register *applicationAuth.Register, login *applicationAuth.Login) *AuthHandler {
	return &AuthHandler{
		register: register,
		login:    login,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	// bind request body
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid request body"))
	}

	// value object conversion
	email, err := valueobject.NewEmail(req.Email)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid email"))
	}
	password, err := valueobject.NewPassword(req.Password)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid password"))
	}
	output, err := h.register.Do(c.Request().Context(), applicationAuth.RegisterInput{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return RespondError(c, err)
	}

	// return response
	return c.JSON(http.StatusCreated, authResponse{
		AccessToken: output.AccessToken,
		User:        toUserResponse(output.User),
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	// bind request body
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid request body"))
	}

	// value object conversion
	email, err := valueobject.NewEmail(req.Email)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid email"))
	}
	password, err := valueobject.NewPassword(req.Password)
	if err != nil {
		return RespondError(c, apperror.NewBadRequestError().SetMessage("invalid password"))
	}
	output, err := h.login.Do(c.Request().Context(), applicationAuth.LoginInput{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return RespondError(c, err)
	}

	// return response
	return c.JSON(http.StatusOK, authResponse{
		AccessToken: output.AccessToken,
		User:        toUserResponse(output.User),
	})
}
