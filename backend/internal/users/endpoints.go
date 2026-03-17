package users

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/errors"
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
		if stderrors.Is(err, ErrUserNotFound) {
			return errors.NewNotFound("user not found")
		}
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, userToResponse(user))
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
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, GetUsersResponse{
		Users: usersToResponse(users),
		Pages: pages,
	})
}

func (h *Handler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return errors.NewBadRequest("validation failed")
	}

	newUser, err := toUserEntity(req)
	if err != nil {
		return errors.NewBadRequest("invalid user data")
	}

	userID, err := h.svc.CreateUser(ctx, newUser)
	if err != nil {
		if stderrors.Is(err, ErrUserAlreadyExists) {
			return errors.NewConflict("user already exists")
		}
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateUsersResponse{
		ID: fmt.Sprintf("%d", userID),
	})
}

func (h *Handler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	newUser := toUserUpdateEntity(req)
	err := h.svc.UpdateUser(ctx, req.ID, newUser)
	if err != nil {
		if stderrors.Is(err, ErrUserNotFound) {
			return errors.NewNotFound("user not found")
		}
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteUserRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	err := h.svc.DeleteUser(ctx, req.ID)
	if err != nil {
		if stderrors.Is(err, ErrUserNotFound) {
			return errors.NewNotFound("user not found")
		}
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}
