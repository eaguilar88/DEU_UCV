package users

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/labstack/echo/v4"
)

type Service interface {
	GetUser(ctx context.Context, userID string) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int64, error)
	UpdateUser(ctx context.Context, userID int, user entities.User) error
	DeleteUser(ctx context.Context, userID int) error
}

type UserEndpointsHandler struct {
	svc Service
	log log.Logger
}

func MakeUserEndpointsHandler(svc Service, log log.Logger) UserEndpointsHandler {
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
		level.Error(h.log).Log("message", "could not decode", "error", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, entitiesUserToGetUserResponse(user))
}

func (h *UserEndpointsHandler) GetUsers(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	scope.GetPageFromVars(c.QueryParam("page"))
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	req := GetUsersRequest{
		PageScope: scope,
	}
	users, pages, err := h.svc.GetUsers(ctx, req.PageScope)
	if err != nil {
		level.Error(h.log).Log("message", "could not decode", "error", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, GetUsersResponse{
		Users: userEntitiesToUserDTO(users),
		Pages: pages,
	})
}

func (h *UserEndpointsHandler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		level.Error(h.log).Log("message", "could not decode", "request", req)
		return echo.ErrBadRequest
	}

	newUser, err := createUserRequestToEntitiesUser(req)
	if err != nil {
		level.Error(h.log).Log("message", "errors creating request", "error", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	userID, err := h.svc.CreateUser(ctx, newUser)
	if err != nil {
		level.Error(h.log).Log("message", "could not decode", "error", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, CreateUsersResponse{
		ID: fmt.Sprintf("%d", userID),
	})
}

func (h *UserEndpointsHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		level.Error(h.log).Log("message", "could not decode", "request", req)
		return c.JSON(http.StatusInternalServerError, err)
	}

	intID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		level.Error(h.log).Log("message", "could not decode", "request", req)
		return c.JSON(http.StatusInternalServerError, err)
	}

	newUser := updateUserRequestToEntitiesUser(req, intID)
	err = h.svc.UpdateUser(ctx, intID, newUser)
	if err != nil {
		level.Error(h.log).Log("message", "could not decode", "error", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *UserEndpointsHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	req := DeleteUserRequest{ID: c.Param("id")}
	intID, err := strconv.Atoi(req.ID)
	if err != nil {
		level.Error(h.log).Log("message", "could not decode", "request", req)
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = h.svc.DeleteUser(ctx, intID)
	if err != nil {
		level.Error(h.log).Log("message", "could not decode", "error", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusAccepted, nil)
}
