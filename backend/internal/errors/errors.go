package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	errScan            = errors.New("scan error")
	errBadQuery        = errors.New("bad query error")
	errDuplicateEntry  = errors.New("duplicated entry")
	errNotFound        = errors.New("rows not found")
	errInvalidPassword = errors.New("invalid password")
	errInternal        = errors.New("internal error")
	ErrNotFound        = NewNotFoundError(errors.New("resource not found"))
	ErrMissingClaims   = NewCustomError(http.StatusUnauthorized, errors.New("missing claims"))
	ErrBadFaculty      = NewBadRequestError(errors.New("el nombre de la facultad es inválido"))
)

type CustomError interface {
	Error() string
	StatusCode() int
	Unwrap() error
	SafeMessage() string
}

type customError struct {
	Cause       error
	Code        int
	UserMessage string
}

func NewScanError(err error) error {
	return fmt.Errorf("%w: %w", errScan, err)
}

func NewBadQueryError(err error) error {
	return fmt.Errorf("%w: %w", errBadQuery, err)
}

func NewDuplicateEntryError(err error) error {
	return fmt.Errorf("%w: %w", errDuplicateEntry, err)
}

func NewInvalidPasswordError(err error) error {
	return fmt.Errorf("%w: %w", errInvalidPassword, err)
}

func IsInternalErr(err error) bool {
	return errors.Is(err, errInternal)
}

func IsScanErr(err error) bool {
	return errors.Is(err, errScan)
}

func IsBadQueryErr(err error) bool {
	return errors.Is(err, errBadQuery)
}

func IsDuplicateEntryErr(err error) bool {
	return errors.Is(err, errDuplicateEntry)
}

func IsNotFoundError(err error) bool {
	return errors.Is(err, errNotFound)
}

func IsInvalidPasswordErr(err error) bool {
	return errors.Is(err, errInvalidPassword)
}

func (c customError) Error() string {
	return c.Cause.Error()
}

func (c customError) StatusCode() int {
	return c.Code
}

func (c customError) Unwrap() error {
	return c.Cause
}

const internalErrorMessage = "there was an internal error. contact an administrator or try again later"

// SafeMessage returns a safe message for API responses.
// For 5xx errors, it always returns a generic message.
// For 4xx errors, it returns the UserMessage if set, otherwise a default message for the status code.
func (c customError) SafeMessage() string {
	if c.Code >= 500 {
		return internalErrorMessage
	}
	if c.UserMessage != "" {
		return c.UserMessage
	}
	return defaultMessageForCode(c.Code)
}

func defaultMessageForCode(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "invalid request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "resource not found"
	case http.StatusConflict:
		return "resource already exists"
	case http.StatusUnprocessableEntity:
		return "unprocessable entity"
	default:
		return "an error occurred"
	}
}

func (c *customError) MarshalJSON() ([]byte, error) {
	resp := map[string]interface{}{
		"code":    c.StatusCode(),
		"message": c.Cause,
	}
	return json.Marshal(resp)
}

// NewCustomError creates a new custom error with optional user message
func NewCustomError(code int, err error, userMessage ...string) CustomError {
	msg := ""
	if len(userMessage) > 0 {
		msg = userMessage[0]
	}
	return &customError{
		Cause:       err,
		Code:        code,
		UserMessage: msg,
	}
}

func NewNotFoundError(err error) CustomError {
	return NewCustomError(http.StatusNotFound, fmt.Errorf("%w: %w", errNotFound, err))
}

func NewUnauthorizedError(err error) CustomError {
	return NewCustomError(http.StatusUnauthorized, err)
}

func NewForbiddenError(err error) CustomError {
	return NewCustomError(http.StatusForbidden, err)
}

func NewInternalError(err error) CustomError {
	return NewCustomError(http.StatusInternalServerError, fmt.Errorf("%w: %w", errInternal, err))
}

func NewUnprocessableError(err error) CustomError {
	return NewCustomError(http.StatusUnprocessableEntity, err)
}

func NewBadRequestError(err error) CustomError {
	return NewCustomError(http.StatusBadRequest, err)
}

// NewNotFound creates a not found error with a safe user message
func NewNotFound(userMessage string) CustomError {
	return &customError{
		Cause:       errNotFound,
		Code:        http.StatusNotFound,
		UserMessage: userMessage,
	}
}

// NewBadRequest creates a bad request error with a safe user message
func NewBadRequest(userMessage string) CustomError {
	return &customError{
		Cause:       errors.New(userMessage),
		Code:        http.StatusBadRequest,
		UserMessage: userMessage,
	}
}

// NewUnauthorized creates an unauthorized error with a safe user message
func NewUnauthorized(userMessage string) CustomError {
	return &customError{
		Cause:       errors.New(userMessage),
		Code:        http.StatusUnauthorized,
		UserMessage: userMessage,
	}
}

// NewForbidden creates a forbidden error with a safe user message
func NewForbidden(userMessage string) CustomError {
	return &customError{
		Cause:       errors.New(userMessage),
		Code:        http.StatusForbidden,
		UserMessage: userMessage,
	}
}

// NewInternal creates an internal server error. Always uses generic message for API response.
func NewInternal(cause error) CustomError {
	return &customError{
		Cause:       fmt.Errorf("%w: %w", errInternal, cause),
		Code:        http.StatusInternalServerError,
		UserMessage: "", // SafeMessage() will return generic message for 500s
	}
}

// NewConflict creates a conflict error with a safe user message
func NewConflict(userMessage string) CustomError {
	return &customError{
		Cause:       errors.New(userMessage),
		Code:        http.StatusConflict,
		UserMessage: userMessage,
	}
}
