package auth

import (
	"context"
	"net/http"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/labstack/echo/v4"
)

type Service interface {
	Login(ctx context.Context, username, password string) (string, *entities.User, error)
}

type AuthEndpointsHandler struct {
	svc Service
	log log.Logger
}

func MakeAuthEndpointsHandler(svc Service, log log.Logger) AuthEndpointsHandler {
	return AuthEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h AuthEndpointsHandler) LoginHandleHTTP(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "error decoding request")
	}
	token, user, err := h.svc.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		level.Error(h.log).Log("message", "error logging user", "error", err)
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
