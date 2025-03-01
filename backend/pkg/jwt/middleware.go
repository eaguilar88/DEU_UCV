package jwt

import (
	"fmt"
	"strings"

	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/labstack/echo/v4"
	// Updated import
)

// JWTMiddleware checks for the existence and validity of a JWT in the context.
func JWTMiddleware(signer Signer, log log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")

			if tokenString == "" {
				level.Error(log).Log("message", "authorization token is missing")
				return echo.NewHTTPError(http.StatusUnauthorized, "authorization token is missing")
			}

			// Remove "Bearer " prefix if present
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")

			// Use your Signer to validate the token
			err := signer.ValidateToken(tokenString)
			if err != nil {
				level.Error(log).Log("message", "invalid authorization token", "error", err)
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("invalid authorization token: %v", err))
			}

			return next(c)
		}
	}
}
