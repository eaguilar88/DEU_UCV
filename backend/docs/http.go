package docs

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

//go:embed swagger
var swaggerFS embed.FS

func DocsHandler(r *mux.Router, docsRoute string, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Serve static files from embedded file system
		swaggerFS, err := fs.Sub(swaggerFS, "swagger")
		if err != nil {
			http.Error(w, "Failed to access embedded file system", http.StatusInternalServerError)
			return
		}

		// Render index.html
		indexHTML, err := fs.ReadFile(swaggerFS, "index.html")
		if err != nil {
			http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
			return
		}

		// Serve index.html with embedded service.yaml
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(indexHTML); err != nil {
			logger.Error("failed to write index.html", zap.Error(err))
		}
	}
}
