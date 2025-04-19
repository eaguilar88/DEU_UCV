package auth

import (
	"context"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	Login(ctx context.Context, username, password string) (string, *entities.User, error)
}

type AuthEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeAuthEndpointsHandler(svc Service, log *zap.Logger) AuthEndpointsHandler {
	return AuthEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h AuthEndpointsHandler) LoginHandleHTTP(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrUnauthorized
	}
	token, user, err := h.svc.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		h.log.Error("error logging user", zap.Error(err))
		return echo.ErrUnauthorized
	}

	response := LoginResponse{
		Token: token,
		User: LoginUserResponse{
			ID:    user.ID,
			Name:  user.FirstName + " " + user.LastName,
			Roles: user.Roles,
		},
	}
	return c.JSON(http.StatusOK, response)
}
