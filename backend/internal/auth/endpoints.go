package auth

import (
	"context"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	Login(ctx context.Context, username, password string) (string, *entities.User, error)
}

type Handler struct {
	svc Service
	log *zap.Logger
}

func NewHandler(svc Service, log *zap.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

func (h *Handler) LoginHandleHTTP(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewUnauthorized("invalid credentials")
	}
	token, user, err := h.svc.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		// Log full error internally, but return generic message to client
		h.log.Error("login failed", zap.Error(err), zap.String("username", req.Username))
		return httperrors.NewUnauthorized("invalid credentials")
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
