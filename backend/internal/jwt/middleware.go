package jwt

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const authErrorMessage = "invalid or expired authorization token"

// JWTMiddleware checks for the existence and validity of a JWT in the context.
func JWTMiddleware(signer Signer, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")

			if tokenString == "" {
				logger.Error("authorization token is missing")
				return echo.NewHTTPError(http.StatusUnauthorized, authErrorMessage)
			}

			// Remove "Bearer " prefix if present
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")

			// Use your Signer to validate the token
			claims, err := signer.ValidateToken(tokenString)
			if err != nil {
				logger.Error("invalid authorization token", zap.Error(err))
				return echo.NewHTTPError(http.StatusUnauthorized, authErrorMessage)
			}

			// Extract "v1" map from claims
			v1Claims, ok := claims["v1"].(map[string]interface{})
			if !ok {
				logger.Error("v1 claims missing or invalid")
				return echo.NewHTTPError(http.StatusUnauthorized, authErrorMessage)
			}

			// Extract userID from v1 map
			userID, ok := v1Claims["userID"].(string)
			if !ok || userID == "" {
				logger.Error("userID missing in v1 claims")
				return echo.NewHTTPError(http.StatusUnauthorized, authErrorMessage)
			}

			// Extract roles from v1 map
			var roles []string
			if rolesRaw, ok := v1Claims["roles"].([]interface{}); ok {
				for _, r := range rolesRaw {
					if roleName, ok := r.(string); ok {
						roles = append(roles, roleName)
					}
				}
			}

			domainType, _ := v1Claims["domainType"].(string)
			faculty, _ := v1Claims["faculty"].(string)
			providerCode, _ := v1Claims["providerCode"].(string)

			c.Set("userID", userID)
			c.Set("roles", roles)
			c.Set("domainType", domainType)
			c.Set("faculty", faculty)
			c.Set("providerCode", providerCode)

			return next(c)
		}
	}
}
