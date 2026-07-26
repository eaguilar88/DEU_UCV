package jwt

import (
	"net/http"
	"strings"

	"github.com/eaguilar88/deu/internal/httperrors"
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
			v1Claims, ok := claims["v1"].(map[string]any)
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
			if rolesRaw, ok := v1Claims["roles"].([]any); ok {
				for _, r := range rolesRaw {
					if roleName, ok := r.(string); ok {
						roles = append(roles, roleName)
					}
				}
			}

			var domainType string
			if d, ok := v1Claims["domainType"].(string); ok {
				domainType = d
			}
			var faculty string
			if f, ok := v1Claims["faculty"].(string); ok {
				faculty = f
			}
			var providerCode string
			if p, ok := v1Claims["providerCode"].(string); ok {
				providerCode = p
			}

			c.Set("userID", userID)
			c.Set("roles", roles)
			c.Set("domainType", domainType)
			c.Set("faculty", faculty)
			c.Set("providerCode", providerCode)

			return next(c)
		}
	}
}

// RequireRoles returns middleware that only allows the request through if
// the authenticated user's roles (set in context by JWTMiddleware) include
// at least one of allowedRoles. It must be chained after JWTMiddleware.
func RequireRoles(allowedRoles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roles, ok := c.Get("roles").([]string)
			if !ok {
				return httperrors.NewForbidden("you do not have permission to access this resource")
			}
			for _, r := range roles {
				if _, ok := allowed[r]; ok {
					return next(c)
				}
			}
			return httperrors.NewForbidden("you do not have permission to access this resource")
		}
	}
}
