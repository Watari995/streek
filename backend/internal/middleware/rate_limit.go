package middleware

import (
	"net/http"

	"github.com/Watari995/streek/backend/internal/apperror"
	"github.com/Watari995/streek/backend/internal/domain/ratelimit"
	"github.com/labstack/echo/v4"
)

func RateLimitMiddleware(limiter ratelimit.IRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := GetUserID(c.Request().Context())
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
			}

			allowed, err := limiter.Allow(c.Request().Context(), userID.String())
			if err != nil {
				c.Logger().Error(err)
				return next(c)
			}

			if !allowed {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]interface{}{
						"code":    string(apperror.CodeRateLimitExceeded),
						"message": "too many requests, please try again later",
					},
				})
			}

			return next(c)
		}
	}
}
