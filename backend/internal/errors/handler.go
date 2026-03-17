package errors

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ErrorResponse is the standard error response format for the API
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewHTTPErrorHandler creates a global Echo error handler that:
// 1. Logs full error details internally using Zap
// 2. Returns safe, human-readable messages to clients
// 3. Never exposes internal error details in API responses
func NewHTTPErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		code := http.StatusInternalServerError
		message := internalErrorMessage

		// Handle CustomError (our domain errors)
		var customErr CustomError
		if errors.As(err, &customErr) {
			code = customErr.StatusCode()
			message = customErr.SafeMessage()
			logger.Error("request error",
				zap.Int("status", code),
				zap.String("safe_message", message),
				zap.Error(customErr),
				zap.String("path", c.Request().URL.Path),
				zap.String("method", c.Request().Method),
			)
		} else {
			// Handle echo.HTTPError
			var he *echo.HTTPError
			if errors.As(err, &he) {
				code = he.Code
				if code < 500 {
					// For 4xx errors from Echo, use the message if it's a string
					if msg, ok := he.Message.(string); ok {
						message = msg
					} else {
						message = defaultMessageForCode(code)
					}
				}
				// For 5xx, message remains the generic internal error message
				logger.Error("HTTP error",
					zap.Int("status", code),
					zap.String("safe_message", message),
					zap.Any("original_message", he.Message),
					zap.Error(he.Internal),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
				)
			} else {
				// Unknown error type - treat as internal error
				logger.Error("unexpected error",
					zap.Int("status", code),
					zap.String("safe_message", message),
					zap.Error(err),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
				)
			}
		}

		// Send error response
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, ErrorResponse{Error: message})
		}

		if err != nil {
			logger.Error("failed to send error response", zap.Error(err))
		}
	}
}
