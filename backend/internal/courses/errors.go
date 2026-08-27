package courses

import "errors"

var (
	ErrCourseNotFound              = errors.New("course not found")
	ErrCourseAlreadyExists         = errors.New("course already exists")
	ErrInvalidInput                = errors.New("invalid input")
	ErrCourseClosureRequestPending = errors.New("course has a pending closure request")
)
