package group_resource_requests_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/group_resource_requests"
	"github.com/eaguilar88/deu/internal/group_resource_requests/mocks"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHandler_RejectGroupResourceRequest(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		prepare    func(svc *mocks.MockService)
		wantStatus int // 0 means no error
	}{
		{
			name: "valid razon is trimmed and passed to the service",
			body: `{"razon": "  Documentación incompleta  "}`,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RejectGroupResourceRequest(mock.Anything, "7", "Documentación incompleta").Return(nil)
			},
		},
		{
			name:       "missing razon returns 400",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "blank razon returns 400",
			body:       `{"razon": "   "}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body returns 400",
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "request not found returns 404",
			body: `{"razon": "Documentación incompleta"}`,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RejectGroupResourceRequest(mock.Anything, "7", "Documentación incompleta").Return(group_resource_requests.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.prepare != nil {
				tt.prepare(svc)
			}

			e := echo.New()
			e.Validator = security.NewCustomValidator()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues("7")

			err := group_resource_requests.NewHandler(svc, zap.NewNop()).RejectGroupResourceRequest(c)
			if tt.wantStatus == 0 {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusAccepted, rec.Code)
				return
			}
			var customErr httperrors.CustomError
			if assert.True(t, errors.As(err, &customErr), "expected CustomError, got %v", err) {
				assert.Equal(t, tt.wantStatus, customErr.StatusCode())
			}
		})
	}
}
