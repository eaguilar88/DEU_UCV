package course_requests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/course_requests/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_GetMyCourseRequests(t *testing.T) {
	type testCase struct {
		name    string
		userID  *string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		resp    any
		wantErr error
	}

	userID := "user-1"

	tests := []testCase{
		{
			name:   "success",
			userID: &userID,
			svc:    &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetMyCourseRequests", ctx.Request().Context(), userID, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.CourseRequest{{ID: "1"}}, entities.PageScope{Page: 1, PerPage: 10}, nil)
			},
			resp: GetCourseRequestsResponse{
				CourseRequests: []GetCourseRequestResponse{{ID: "1"}},
				Pages:          entities.PageScope{Page: 1, PerPage: 10},
			},
		},
		{
			name:   "success empty list",
			userID: &userID,
			svc:    &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetMyCourseRequests", ctx.Request().Context(), userID, mock.AnythingOfType("entities.PageScope")).
					Return(nil, entities.PageScope{}, nil)
			},
			resp: GetCourseRequestsResponse{
				CourseRequests: []GetCourseRequestResponse{},
			},
		},
		{
			name:   "error getting my course requests",
			userID: &userID,
			svc:    &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetMyCourseRequests", ctx.Request().Context(), userID, mock.AnythingOfType("entities.PageScope")).
					Return(nil, entities.PageScope{}, errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
		{
			name:    "error cannot find userID in context",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/course-requests", nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.GetMyCourseRequests(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetCourseRequestsResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

// assertCustomError compares errors by status code and safe message.
func assertCustomError(t *testing.T, expected, actual error) {
	t.Helper()
	if expected == nil {
		assert.Nil(t, actual)
		return
	}
	if actual == nil {
		t.Errorf("expected error %v but got nil", expected)
		return
	}
	expectedErr, ok := expected.(httperrors.CustomError)
	if !ok {
		t.Errorf("expected error is not CustomError: %T", expected)
		return
	}
	actualErr, ok := actual.(httperrors.CustomError)
	if !ok {
		t.Errorf("actual error is not CustomError: %T", actual)
		return
	}
	assert.Equal(t, expectedErr.StatusCode(), actualErr.StatusCode(), "status codes should match")
	assert.Equal(t, expectedErr.SafeMessage(), actualErr.SafeMessage(), "safe messages should match")
}
