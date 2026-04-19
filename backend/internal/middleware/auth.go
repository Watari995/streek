package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Watari995/streek/backend/domain/valueobject"
	"github.com/Watari995/streek/backend/infrastructure/auth"
	"github.com/labstack/echo/v4"
)

// key for userID in context
type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// get header
			header := c.Request().Header.Get("Authorization")

			// get token from "Bearer <token>"
			if !strings.HasPrefix(header, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing or invalid authorization header",
				})
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")

			// validate token and get userID
			userID, err := auth.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid token",
				})
			}

			// set userID to context
			ctx := context.WithValue(c.Request().Context(), userIDKey, userID)
			c.SetRequest(c.Request().WithContext(ctx))

			// move on to next handler
			return next(c)
		}
	}
}

// helper function to get userID from ctx
func GetUserID(ctx context.Context) (valueobject.UserID, error) {
	userID, ok := ctx.Value(userIDKey).(valueobject.UserID)
	if !ok {
		return valueobject.UserID{}, errors.New("user ID not found in context")
	}
	return userID, nil
}
