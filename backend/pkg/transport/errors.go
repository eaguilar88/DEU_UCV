package transport

import (
	"context"
	"encoding/json"
	"errors"

	"net/http"
	"strings"

	"github.com/go-kit/log"

	kitHTTP "github.com/go-kit/kit/transport/http"
)

const internalServerBodyError = `{"code":500,"message":"internal server error"}`

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
		"code":    c.Code,
		"message": getGRPCErrorMessageFromError(c.Cause),
	}
	return json.Marshal(resp)
}

// NewCustomError creates a new custom error
func NewCustomError(code int, err string) CustomError {
	return &customError{
		Cause: errors.New(err),
		Code:  code,
	}
}

func NewNotFoundError(err string) CustomError {
	return NewCustomError(http.StatusNotFound, err)
}

func NewUnauthorizedError(err string) CustomError {
	return NewCustomError(http.StatusUnauthorized, err)
}

func NewForbiddenError(err string) CustomError {
	return NewCustomError(http.StatusForbidden, err)
}

func NewInternalError(err string) CustomError {
	return NewCustomError(http.StatusInternalServerError, err)
}

func NewUnprocessableError(err string) CustomError {
	return NewCustomError(http.StatusUnprocessableEntity, err)
}

func NewBadRequestError(err string) CustomError {
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

func getGRPCErrorMessageFromError(err error) string {
	errStr := err.Error()
	if !strings.HasPrefix(errStr, "rpc error: ") {
		return errStr
	}
	details := errStr[11:]
	index := strings.Index(details, "desc = ")
	if index == -1 {
		return details
	}
	message := details[index+7:]
	return message
}
