package files

import (
	"context"
	"io"
	"net/http"

	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type StorageClient interface {
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
}

type Handler struct {
	storage StorageClient
	logger  *zap.Logger
}

func NewHandler(storage StorageClient, logger *zap.Logger) *Handler {
	return &Handler{storage: storage, logger: logger}
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

	ctx := c.Request().Context()
	body, contentType, err := h.storage.GetObject(ctx, key)
	if err != nil {
		h.logger.Error("failed to fetch file from storage", zap.Error(err), zap.String("key", key))
		return httperrors.NewNotFound("file not found")
	}
	defer func() {
		if err := body.Close(); err != nil {
			h.logger.Warn("failed to close file body", zap.Error(err), zap.String("key", key))
		}
	}()

	return c.Stream(http.StatusOK, contentType, body)
}
