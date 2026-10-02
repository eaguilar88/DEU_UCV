package users

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidInput      = errors.New("invalid input")
	// ErrUserRoleNotFound means the user does not hold the role being replaced.
	ErrUserRoleNotFound = errors.New("user role not found")
)
