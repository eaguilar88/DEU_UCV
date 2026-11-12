package users

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/jwt"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetUser(ctx context.Context, userID string, viewer entities.Viewer) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User, profilePic *entities.File) (int64, error)
	UpdateUser(ctx context.Context, userID string, changes entities.UserChanges, profilePic *entities.File, viewer entities.Viewer) (entities.User, error)
	DeleteUser(ctx context.Context, userID string, viewer entities.Viewer) error
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
	user, err := h.svc.GetUser(ctx, req.ID, jwt.ViewerFromContext(c))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return httperrors.NewNotFound("user not found")
		}
		return httperrors.NewInternal(err)
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
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, GetUsersResponse{
		Users: usersToResponse(users),
		Pages: pages,
	})
}

func (h *Handler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	req := CreateUserRequest{
		Document:       c.FormValue("cedula"),
		Email:          c.FormValue("email"),
		FirstName:      c.FormValue("nombres"),
		LastName:       c.FormValue("apellidos"),
		DateOfBirth:    c.FormValue("fecha_de_nacimiento"),
		Gender:         c.FormValue("genero"),
		EducationLevel: c.FormValue("nivel_educativo"),
		Address:        c.FormValue("direccion"),
		Password:       c.FormValue("password"),
	}

	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	newUser, err := toUserEntity(req)
	if err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	profilePic, err := utils.GetFileFrom(c, "profile_picture")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return httperrors.NewBadRequest("invalid profile picture")
	}

	userID, err := h.svc.CreateUser(ctx, newUser, profilePic)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return httperrors.NewConflict("user already exists")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateUsersResponse{
		ID: fmt.Sprintf("%d", userID),
	})
}

// UpdateUser applies a partial update from a multipart/form-data body: only the fields sent
// are changed, and a profile_picture file replaces the current one.
func (h *Handler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()

	// FormParams parses the body; PostForm then keeps only the body fields, without the
	// query string.
	if _, err := c.FormParams(); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	form := c.Request().PostForm
	formValue := func(key string) *string {
		values, ok := form[key]
		if !ok || len(values) == 0 {
			return nil
		}
		return &values[0]
	}
	req := UpdateUserRequest{
		ID:             c.Param("id"),
		Document:       formValue("cedula"),
		Email:          formValue("email"),
		FirstName:      formValue("nombres"),
		LastName:       formValue("apellidos"),
		DateOfBirth:    formValue("fecha_de_nacimiento"),
		Gender:         formValue("genero"),
		EducationLevel: formValue("nivel_educativo"),
		Address:        formValue("direccion"),
	}

	changes, err := toUserChanges(req)
	if err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	profilePic, err := utils.GetFileFrom(c, "profile_picture")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return httperrors.NewBadRequest("invalid profile picture")
	}

	user, err := h.svc.UpdateUser(ctx, req.ID, changes, profilePic, jwt.ViewerFromContext(c))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return httperrors.NewBadRequest("no fields to update")
		case errors.Is(err, ErrUserForbidden):
			return httperrors.NewForbidden("not allowed to modify this user")
		case errors.Is(err, ErrUserNotFound):
			return httperrors.NewNotFound("user not found")
		case errors.Is(err, ErrUserAlreadyExists):
			return httperrors.NewConflict("email or cedula already in use")
		}
		h.log.Error("failed to update user", zap.Error(err), zap.String("user_id", req.ID))
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, userToResponse(user))
}

func (h *Handler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteUserRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}

	err := h.svc.DeleteUser(ctx, req.ID, jwt.ViewerFromContext(c))
	if err != nil {
		if errors.Is(err, ErrUserForbidden) {
			return httperrors.NewForbidden("not allowed to delete this user")
		}
		if errors.Is(err, ErrUserNotFound) {
			return httperrors.NewNotFound("user not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}
