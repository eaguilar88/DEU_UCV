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
	"go.uber.org/zap"
)

func TestHandler_ServeFile(t *testing.T) {
	const key = "files/activities/24/cubierta.png"

	tests := []struct {
		name       string
		path       string
		prepare    func(svc *mocks.MockService)
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{
			name: "public file is streamed with cache header",
			path: "/files/" + key,
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
			path: "/files/" + key,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().GetPublicFile(mock.Anything, key).Return(nil, "", ErrFileNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing key is a 400",
			path:       "/files/",
			prepare:    func(_ *mocks.MockService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			tt.prepare(svc)

			e := echo.New()
			e.HTTPErrorHandler = httperrors.NewHTTPErrorHandler(zap.NewNop())
			e.GET("/files/*", NewHandler(svc, zap.NewNop()).ServeFile)

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, tt.wantBody, rec.Body.String())
				assert.Equal(t, "image/png", rec.Header().Get(echo.HeaderContentType))
				assert.Equal(t, tt.wantCache, rec.Header().Get(echo.HeaderCacheControl))
			}
		})
	}
}
