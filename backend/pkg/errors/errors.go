package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kit/log"

	kitHTTP "github.com/go-kit/kit/transport/http"
)

const internalServerBodyError = `{"code":500,"message":"internal server error"}`

var (
	errScan            = errors.New("scan error")
	errBadQuery        = errors.New("bad query error")
	errDuplicateEntry  = errors.New("duplicated entry")
	errNotFound        = errors.New("rows not found")
	errInvalidPassword = errors.New("invalid password")
	errInternal        = errors.New("internal error")
)

type CustomError interface {
	Error() string
	StatusCode() int
	Unwrap() error
}

type customError struct {
	Cause error
	Code  int
}

type httpError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
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

func (c *customError) MarshalJSON() ([]byte, error) {
	resp := map[string]interface{}{
		"code":    c.StatusCode(),
		"message": c.Cause,
	}
	return json.Marshal(resp)
}

// NewCustomError creates a new custom error
func NewCustomError(code int, err error) CustomError {
	return &customError{
		Cause: err,
		Code:  code,
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

// MakeHTTPErrorEncoder
func MakeHTTPErrorEncoder(logger log.Logger) kitHTTP.ErrorEncoder {
	return func(_ context.Context, err error, w http.ResponseWriter) {
		logger.Log("error", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if headerer, ok := err.(kitHTTP.Headerer); ok {
			for k, values := range headerer.Headers() {
				for _, v := range values {
					w.Header().Add(k, v)
				}
			}
		}
		b := []byte(internalServerBodyError)
		code := http.StatusInternalServerError
		if sc, ok := err.(kitHTTP.StatusCoder); ok {
			code = sc.StatusCode()
		}
		if ce, ok := err.(*customError); ok {
			code = ce.StatusCode()
			if ce.StatusCode() != http.StatusInternalServerError {
				b, _ = ce.MarshalJSON()
			}
		}

		w.WriteHeader(code)
		w.Write(b)
	}
}
