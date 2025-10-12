package jwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type claims struct {
	UserID  string           `json:"userID"`
	Roles   []string         `json:"roles"`
	Faculty entities.Faculty `json:"faculty"`
}

// JWTMiddleware checks for the existence and validity of a JWT in the context.
func JWTMiddleware(signer Signer, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")

			if tokenString == "" {
				logger.Error("authorization token is missing")
				return echo.NewHTTPError(http.StatusUnauthorized, "authorization token is missing")
			}

			// Remove "Bearer " prefix if present
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")

			// Use your Signer to validate the token
			claims, err := signer.ValidateToken(tokenString)
			if err != nil {
				logger.Error("invalid authorization token", zap.Error(err))
				return echo.NewHTTPError(
					http.StatusUnauthorized,
					fmt.Sprintf("invalid authorization token: %v", err),
				)
			}

			// Extract "v1" map from claims
			v1Claims, ok := claims["v1"].(map[string]interface{})
			if !ok {
				logger.Error("v1 claims missing or invalid")
				return echo.NewHTTPError(http.StatusUnauthorized, "v1 claims missing or invalid")
			}

			// Extract userID from v1 map
			userID, ok := v1Claims["userID"].(string)
			if !ok || userID == "" {
				logger.Error("userID missing in v1 claims")
				return echo.NewHTTPError(http.StatusUnauthorized, "userID missing in v1 claims")
			}

			// Store userID into the context for downstream handlers
			c.Set("userID", userID)

			return next(c)
		}
	}
}
