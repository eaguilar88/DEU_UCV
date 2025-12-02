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

type UserEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeUserEndpointsHandler(svc Service, log *zap.Logger) UserEndpointsHandler {
	return UserEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h *UserEndpointsHandler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetUserRequest{ID: c.Param("id")}
	user, err := h.svc.GetUser(ctx, req.ID)
	if err != nil {
		h.log.Error(fmt.Sprintf("error getting user with ID: %s", req.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, UserEntityToGetUserResponse(user))
}

func (h *UserEndpointsHandler) GetUsers(c echo.Context) error {
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
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, GetUsersResponse{
		Users: UserEntitiesToGetUserResponse(users),
		Pages: pages,
	})
}

func (h *UserEndpointsHandler) CreateUser(c echo.Context) error {
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
		return c.JSON(http.StatusBadRequest, err)
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

func (h *UserEndpointsHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error decoding request", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	newUser := updateUserRequestToEntitiesUser(req)
	err := h.svc.UpdateUser(ctx, req.ID, newUser)
	if err != nil {
		h.log.Error(fmt.Sprintf("error updating user with ID: %s", req.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *UserEndpointsHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteUserRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error decoding request", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	err := h.svc.DeleteUser(ctx, req.ID)
	if err != nil {
		h.log.Error(fmt.Sprintf("error deleting user with ID: %s", req.ID), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *UserEndpointsHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return c.JSON(http.StatusNotFound, err)
	case errors.Is(err, ErrUserAlreadyExists):
		return c.JSON(http.StatusConflict, err)
	default:
		return c.JSON(http.StatusInternalServerError, err)
	}
}
