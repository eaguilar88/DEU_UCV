package files

import (
	"context"
	"io"
	"net/http"

	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// publicCacheControl lets browsers and proxies cache public images (logos, covers) for a day.
const publicCacheControl = "public, max-age=86400"

type Service interface {
	GetPublicFile(ctx context.Context, key string) (io.ReadCloser, string, error)
}

type Handler struct {
	svc    Service
	logger *zap.Logger
}

func NewHandler(svc Service, logger *zap.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// KeyParam is the name Echo gives the wildcard segment of "/files/*". The object key is
// everything after "/files/", e.g. "/files/files/activities/24/cubierta.png" serves the
// object "files/activities/24/cubierta.png".
const KeyParam = "*"

func (h *Handler) ServeFile(c echo.Context) error {
	key := c.Param(KeyParam)
	if key == "" {
		return httperrors.NewBadRequest("missing file key")
	}

	body, contentType, err := h.svc.GetPublicFile(c.Request().Context(), key)
	if err != nil {
		h.logger.Debug("file not served", zap.Error(err), zap.String("key", key))
		return httperrors.NewNotFound("file not found")
	}
	defer func() {
		if err := body.Close(); err != nil {
			h.logger.Warn("failed to close file body", zap.Error(err), zap.String("key", key))
		}
	}()

	c.Response().Header().Set(echo.HeaderCacheControl, publicCacheControl)
	return c.Stream(http.StatusOK, contentType, body)
}
