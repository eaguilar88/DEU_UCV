package docs

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

//go:embed swagger
var swaggerFS embed.FS

func RegisterDocsRoute(e *echo.Echo, logger *zap.Logger) {
	subFS, err := fs.Sub(swaggerFS, "swagger")
	if err != nil {
		logger.Fatal("failed to create swagger sub-filesystem", zap.Error(err))
	}
	assetHandler := http.FileServer(http.FS(subFS))
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/")
	})
	e.GET("/docs/*", echo.WrapHandler(http.StripPrefix("/docs/", assetHandler)))
}
