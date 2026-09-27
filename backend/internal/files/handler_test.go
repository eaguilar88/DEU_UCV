package files

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/files/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// TestServeFile_Routing goes through a real Echo router, so it catches a mismatch between
// the route pattern and the param name the handler reads.
func TestServeFile_Routing(t *testing.T) {
	storage := mocks.NewMockStorageClient(t)
	storage.EXPECT().GetObject(mock.Anything, "files/activities/24/cubierta.png").
		Return(io.NopCloser(strings.NewReader("png")), "image/png", nil)

	e := echo.New()
	e.GET("/files/*", NewHandler(storage, zap.NewNop()).ServeFile)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/files/files/activities/24/cubierta.png", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "png", rec.Body.String())
	assert.Equal(t, "image/png", rec.Header().Get(echo.HeaderContentType))
}
