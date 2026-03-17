package files

import (
	"context"
	"io"
	"net/http"

	"github.com/eaguilar88/deu/internal/errors"
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

func (h *Handler) ServeFile(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return errors.NewBadRequest("missing file key")
	}

	ctx := c.Request().Context()
	body, contentType, err := h.storage.GetObject(ctx, key)
	if err != nil {
		h.logger.Error("failed to fetch file from storage", zap.Error(err), zap.String("key", key))
		return errors.NewNotFound("file not found")
	}
	defer body.Close()

	return c.Stream(http.StatusOK, contentType, body)
}
