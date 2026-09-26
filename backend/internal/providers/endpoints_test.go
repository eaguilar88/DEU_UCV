package providers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/providers/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_GetProvider(t *testing.T) {
	tests := []struct {
		name       string
		svcErr     error
		wantStatus int
	}{
		{name: "allowed", wantStatus: http.StatusOK},
		{name: "forbidden", svcErr: ErrProviderForbidden, wantStatus: http.StatusForbidden},
		{name: "not found", svcErr: ErrProviderNotFound, wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/providers/5", nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues("5")
			c.Set("userID", "10")

			svc := mocks.NewMockService(t)
			svc.EXPECT().GetProvider(mock.Anything, "5", entities.Viewer{UserID: "10"}).
				Return(entities.Provider{ID: "5"}, tt.svcErr)

			err := NewHandler(svc, zap.NewNop()).GetProvider(c)

			if tt.svcErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatus, rec.Code)
				return
			}
			var customErr httperrors.CustomError
			require.ErrorAs(t, err, &customErr)
			assert.Equal(t, tt.wantStatus, customErr.StatusCode())
		})
	}
}
