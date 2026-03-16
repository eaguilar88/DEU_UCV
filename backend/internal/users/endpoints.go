package users

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetUser(ctx context.Context, userID string) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int64, error)
	UpdateUser(ctx context.Context, userID string, user entities.User) error
	DeleteUser(ctx context.Context, userID string) error
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

func (h *Handler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetUserRequest{ID: c.Param("id")}
	user, err := h.svc.GetUser(ctx, req.ID)
	if err != nil {
		h.log.Error("error getting user", zap.String("id", req.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, UserEntityToGetUserResponse(user))
}

func (h *Handler) GetUsers(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	req := GetUsersRequest{
		PageScope: scope,
	}
	users, pages, err := h.svc.GetUsers(ctx, req.PageScope)
	if err != nil {
		h.log.Error("error getting users", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, GetUsersResponse{
		Users: UserEntitiesToGetUserResponse(users),
		Pages: pages,
	})
}

func (h *Handler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error decoding create user request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		h.log.Error("error validating request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	newUser, err := createUserRequestToEntitiesUser(req)
	if err != nil {
		h.log.Error("error creating new user entity", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	userID, err := h.svc.CreateUser(ctx, newUser)
	if err != nil {
		h.log.Error("error creating new user", zap.Error(err))
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, CreateUsersResponse{
		ID: fmt.Sprintf("%d", userID),
	})
}

func (h *Handler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error decoding request", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	newUser := updateUserRequestToEntitiesUser(req)
	err := h.svc.UpdateUser(ctx, req.ID, newUser)
	if err != nil {
		h.log.Error("error updating user", zap.String("id", req.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteUserRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error decoding request", zap.Error(err))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err := h.svc.DeleteUser(ctx, req.ID)
	if err != nil {
		h.log.Error("error deleting user", zap.String("id", req.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return c.JSON(http.StatusNotFound, err.Error())
	case errors.Is(err, ErrUserAlreadyExists):
		return c.JSON(http.StatusConflict, err.Error())
	default:
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
}
