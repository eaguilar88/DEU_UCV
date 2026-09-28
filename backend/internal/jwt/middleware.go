package jwt

import (
	"errors"
	"fmt"
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
			if err := authenticate(c, signer); err != nil {
				logger.Error("authentication failed", zap.Error(err))
				return echo.NewHTTPError(http.StatusUnauthorized, authErrorMessage)
			}
			return next(c)
		}
	}
}

// OptionalJWTMiddleware authenticates the request when a valid token is present and
// otherwise lets it through anonymously. It is meant for public endpoints whose response
// depends on who is asking; handlers must treat a missing "userID" as an anonymous caller.
// A missing, invalid or expired token never fails the request.
func OptionalJWTMiddleware(signer Signer, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get("Authorization") == "" {
				return next(c)
			}
			if err := authenticate(c, signer); err != nil {
				logger.Debug("optional authentication failed, continuing as anonymous", zap.Error(err))
			}
			return next(c)
		}
	}
}

// authenticate validates the request's bearer token and stores its v1 claims in the
// context. Nothing is stored in the context when it returns an error.
func authenticate(c echo.Context, signer Signer) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return errors.New("authorization token is missing")
	}

	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := signer.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid authorization token: %w", err)
	}

	// Extract "v1" map from claims
	v1Claims, ok := claims["v1"].(map[string]any)
	if !ok {
		return errors.New("v1 claims missing or invalid")
	}

	// Extract userID from v1 map
	userID, ok := v1Claims["userID"].(string)
	if !ok || userID == "" {
		return errors.New("userID missing in v1 claims")
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

	c.Set("userID", userID)
	c.Set("roles", roles)
	for _, key := range []string{"domainType", "faculty", "providerCode", "groupID", "groupName", "providerID", "providerName"} {
		var value string
		if v, ok := v1Claims[key].(string); ok {
			value = v
		}
		c.Set(key, value)
	}

	return nil
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
