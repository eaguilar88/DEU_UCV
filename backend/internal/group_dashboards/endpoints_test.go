package group_dashboards_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/group_dashboards"
	"github.com/eaguilar88/deu/internal/group_dashboards/mocks"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHandler_GetGroupDashboard(t *testing.T) {
	tests := []struct {
		name       string
		groupID    string
		prepare    func(svc *mocks.MockService)
		wantStatus int
	}{
		{
			name:    "numeric group id is accepted",
			groupID: "31",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().GetGroupDashboard(mock.Anything, "31").Return(&group_dashboards.GroupDashboardResponse{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-numeric group id is rejected",
			groupID:    "abc",
			wantStatus: http.StatusBadRequest,
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
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("groupId")
			c.SetParamValues(tt.groupID)

			err := group_dashboards.NewHandler(svc, zap.NewNop()).GetGroupDashboard(c)
			if tt.wantStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
				return
			}
			var httpErr *echo.HTTPError
			if assert.True(t, errors.As(err, &httpErr), "expected echo.HTTPError, got %v", err) {
				assert.Equal(t, tt.wantStatus, httpErr.Code)
			}
		})
	}
}
