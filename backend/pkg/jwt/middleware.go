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
			claims, err := signer.ValidateToken(tokenString)
			if err != nil {
				level.Error(log).Log("message", "invalid authorization token", "error", err)
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("invalid authorization token: %v", err))
			}

			// Extract "v1" map from claims
			v1Claims, ok := claims["v1"].(map[string]interface{})
			if !ok {
				level.Error(log).Log("message", "v1 claims missing or invalid")
				return echo.NewHTTPError(http.StatusUnauthorized, "v1 claims missing or invalid")
			}

			// Extract userID from v1 map
			userID, ok := v1Claims["userID"].(string)
			if !ok || userID == "" {
				level.Error(log).Log("message", "userID missing in v1 claims")
				return echo.NewHTTPError(http.StatusUnauthorized, "userID missing in v1 claims")
			}

			// Store userID into the context for downstream handlers
			c.Set("userID", userID)

			return next(c)
		}
	}
}
