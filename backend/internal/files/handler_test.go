package files

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/files/mocks"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_ServeFile(t *testing.T) {
	const key = "files/groups/1/logo_logo.png"

	tests := []struct {
		name        string
		key         string
		prepare     func(svc *mocks.MockService)
		wantStatus  int
		wantBody    string
		wantCache   string
		wantErrCode int
	}{
		{
			name: "public file is streamed with cache header",
			key:  key,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().GetPublicFile(mock.Anything, key).
					Return(io.NopCloser(strings.NewReader("png")), "image/png", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   "png",
			wantCache:  publicCacheControl,
		},
		{
			name: "service error is a 404",
			key:  key,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().GetPublicFile(mock.Anything, key).Return(nil, "", ErrFileNotFound)
			},
			wantErrCode: http.StatusNotFound,
		},
		{
			name:        "missing key is a 400",
			key:         "",
			prepare:     func(_ *mocks.MockService) {},
			wantErrCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			tt.prepare(svc)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/files/"+tt.key, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("key")
			c.SetParamValues(tt.key)

			err := NewHandler(svc, zap.NewNop()).ServeFile(c)

			if tt.wantErrCode != 0 {
				var customErr httperrors.CustomError
				require.ErrorAs(t, err, &customErr)
				assert.Equal(t, tt.wantErrCode, customErr.StatusCode())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantBody, rec.Body.String())
			assert.Equal(t, tt.wantCache, rec.Header().Get(echo.HeaderCacheControl))
		})
	}
}

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
