package auth

type LoginRequest struct {
	Username string `json:"usuario"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"usuario"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"passwordActual" validate:"required"`
	NewPassword     string `json:"passwordNueva" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"passwordNueva" validate:"required,min=8"`
}
