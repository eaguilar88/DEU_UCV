package groups

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups/mocks"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

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

func TestHandler_GetGroups(t *testing.T) {
	type testCase struct {
		name    string
		url     string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "random without limit defaults to 3",
			url:  "/groups?random=true",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetRandomActiveGroups", ctx.Request().Context(), 3).
					Return([]entities.ExtensionGroup{{ID: "1"}, {ID: "2"}, {ID: "3"}}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			},
		},
		{
			name: "random with valid limit",
			url:  "/groups?random=true&limit=5",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetRandomActiveGroups", ctx.Request().Context(), 5).
					Return([]entities.ExtensionGroup{{ID: "1"}}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}},
			},
		},
		{
			name: "random limit clamped to max",
			url:  "/groups?random=true&limit=9999",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetRandomActiveGroups", ctx.Request().Context(), maxRandomGroupsLimit).
					Return([]entities.ExtensionGroup{{ID: "1"}}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}},
			},
		},
		{
			name: "random with invalid limit falls back to default",
			url:  "/groups?random=true&limit=abc",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetRandomActiveGroups", ctx.Request().Context(), defaultRandomGroupsLimit).
					Return([]entities.ExtensionGroup{{ID: "1"}}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}},
			},
		},
		{
			name: "random service error",
			url:  "/groups?random=true",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetRandomActiveGroups", ctx.Request().Context(), defaultRandomGroupsLimit).
					Return(nil, errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
		{
			name: "no random param uses paginated path",
			url:  "/groups",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetGroups", ctx.Request().Context(), entities.GroupFilter{}, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.ExtensionGroup{{ID: "1"}}, entities.PageScope{Page: 1, PerPage: 10}, nil)
			},
			resp: GetGroupsResponse{
				Groups:    []GetGroupResponse{{ID: "1"}},
				PageScope: entities.PageScope{Page: 1, PerPage: 10},
			},
		},
		{
			name: "faculty filter passed through",
			url:  "/groups?faculty=Ingenier%C3%ADa",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetGroups", ctx.Request().Context(), entities.GroupFilter{Faculty: entities.FacultyIngenieria}, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.ExtensionGroup{{ID: "1"}}, entities.PageScope{}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}},
			},
		},
		{
			name:    "invalid faculty returns bad request",
			url:     "/groups?faculty=NotARealFaculty",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewBadRequest("invalid faculty: NotARealFaculty"),
		},
		{
			name: "type filter passed through",
			url:  "/groups?type=cultural",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetGroups", ctx.Request().Context(), entities.GroupFilter{Type: entities.CulturalGroupType}, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.ExtensionGroup{{ID: "1"}}, entities.PageScope{}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}},
			},
		},
		{
			name: "active filter passed through",
			url:  "/groups?active=true",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				active := true
				tc.svc.On("GetGroups", ctx.Request().Context(), entities.GroupFilter{Active: &active}, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.ExtensionGroup{{ID: "1"}}, entities.PageScope{}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}},
			},
		},
		{
			name:    "invalid active returns bad request",
			url:     "/groups?active=notabool",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewBadRequest("active must be true or false"),
		},
		{
			name: "deleted filter passed through",
			url:  "/groups?deleted=true",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetGroups", ctx.Request().Context(), entities.GroupFilter{Deleted: true}, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.ExtensionGroup{{ID: "1"}}, entities.PageScope{}, nil)
			},
			resp: GetGroupsResponse{
				Groups: []GetGroupResponse{{ID: "1"}},
			},
		},
		{
			name:    "invalid deleted returns bad request",
			url:     "/groups?deleted=notabool",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewBadRequest("deleted must be true or false"),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.GetGroups(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetGroupsResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
		})
	}
}
