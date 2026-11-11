package users

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidInput      = errors.New("invalid input")
	// ErrUserForbidden means the caller may not modify or delete this user.
	ErrUserForbidden = errors.New("not allowed to modify this user")
	// ErrUserRoleNotFound means the user does not hold the role being replaced.
	ErrUserRoleNotFound = errors.New("user role not found")
)
