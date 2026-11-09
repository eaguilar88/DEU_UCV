package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
)

type Service interface {
	Login(ctx context.Context, username, password string) (string, *entities.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// forgotPasswordMessage is returned whether or not the email is registered.
const forgotPasswordMessage = "Si el correo está registrado, recibirás un enlace para restablecer tu contraseña."

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) LoginHandleHTTP(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewUnauthorized("invalid credentials")
	}
	token, user, err := h.svc.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return httperrors.NewUnauthorized("invalid credentials")
	}

	response := LoginResponse{
		Token: token,
		User: LoginUserResponse{
			ID:    user.ID,
			Name:  user.FirstName + " " + user.LastName,
			Roles: user.Roles,
			Rol:   entities.UIRole(user.Roles),
		},
		Faculty:            user.Faculty,
		ProviderCode:       user.ProviderCode,
		GroupID:            user.GroupID,
		GroupName:          user.GroupName,
		CourseProviderID:   user.CourseProviderID,
		CourseProviderName: user.CourseProviderName,
	}
	return c.JSON(http.StatusOK, response)
}

// ChangePasswordHandleHTTP changes the authenticated user's password.
func (h *Handler) ChangePasswordHandleHTTP(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	if err := h.svc.ChangePassword(c.Request().Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidCurrentPassword) {
			return httperrors.NewUnauthorized("contraseña actual incorrecta")
		}
		return httperrors.NewInternal(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// ForgotPasswordHandleHTTP sends a password reset link. It answers the same way whether or not
// the email is registered.
func (h *Handler) ForgotPasswordHandleHTTP(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	if err := h.svc.RequestPasswordReset(c.Request().Context(), req.Email); err != nil {
		return httperrors.NewInternal(err)
	}
	return c.JSON(http.StatusOK, ForgotPasswordResponse{Message: forgotPasswordMessage})
}

// ResetPasswordHandleHTTP sets a new password using the token from a password reset link.
func (h *Handler) ResetPasswordHandleHTTP(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	if err := h.svc.ResetPassword(c.Request().Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidResetToken) {
			return httperrors.NewBadRequest("enlace inválido o expirado")
		}
		return httperrors.NewInternal(err)
	}
	return c.NoContent(http.StatusNoContent)
}
