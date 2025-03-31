package course_periods

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/pkg/course_periods/mocks"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMakeCoursePeriodEndpointsHandler(t *testing.T) {
	type testCase struct {
		name string
		svc  Service
		log  *zap.Logger
		want CoursePeriodEndpointsHandler
	}
	tc := testCase{
		name: "success",
		svc:  &mocks.ServiceMock{},
		log:  zap.NewNop(),
		want: CoursePeriodEndpointsHandler{
			svc: &mocks.ServiceMock{},
			log: zap.NewNop(),
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := MakeCoursePeriodEndpointsHandler(tc.svc, tc.log)
		assert.Equal(t, tc.want.svc, got.svc)
	})
}

func TestCoursePeriodEndpointsHandler_GetCoursePeriod(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.ServiceMock
		prepare func(ctx echo.Context, tc *testCase)
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.CoursePeriod{ID: "1"}, tc.wantErr)
			},
			req: GetCoursePeriodRequest{},
			resp: GetCoursePeriodResponse{
				ID: "1",
			},
			wantErr: nil,
		},
		{
			name: "error getting course period",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.CoursePeriod{}, errors.New("cannot get course period"))
			},
			req:     GetCoursePeriodRequest{},
			resp:    GetCoursePeriodResponse{},
			wantErr: echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.ServiceMock{},
			req:     "bad request",
			wantErr: echo.NewHTTPError(http.StatusBadRequest, "Bad Request"),
		},
	}

	loggerMock := zap.NewNop()
	path := "/courses/:course_id/course_periods/:id"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodGet, path, bytes.NewBuffer(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := MakeCoursePeriodEndpointsHandler(tt.svc, loggerMock)
			err = h.GetCoursePeriod(ctx)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetCoursePeriodResponse
				err = json.Unmarshal(rec.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestCoursePeriodEndpointsHandler_GetCoursePeriods(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.ServiceMock
		prepare func(ctx echo.Context, tc *testCase)
		req     any
		resp    any
		wantErr error
	}
	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCoursePeriods", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.PageScope")).
					Return([]entities.CoursePeriod{{ID: "1"}}, entities.PageScope{}, nil)
			},
			req: GetCoursePeriodRequest{},
			resp: GetCoursePeriodsResponse{
				Periods: []GetCoursePeriodResponse{
					{ID: "1"},
				},
				Pages: entities.PageScope{},
			},
			wantErr: nil,
		},
		{
			name: "error getting course periods",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCoursePeriods", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.PageScope")).
					Return([]entities.CoursePeriod{}, entities.PageScope{}, echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"))
			},
			req:     GetCoursePeriodRequest{},
			resp:    GetCoursePeriodResponse{},
			wantErr: echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"),
		},
	}

	loggerMock := zap.NewNop()
	path := "/courses/:course_id/course_periods"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodGet, path, bytes.NewBuffer(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := MakeCoursePeriodEndpointsHandler(tt.svc, loggerMock)
			err = h.GetCoursePeriods(ctx)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetCoursePeriodsResponse
				err = json.Unmarshal(rec.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestCoursePeriodEndpointsHandler_CreateCoursePeriod(t *testing.T) {
	var userID = "1"
	type testCase struct {
		name    string
		svc     *mocks.ServiceMock
		prepare func(ctx echo.Context, tc *testCase)
		userID  *string
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(1), nil)
			},
			userID: &userID,
			req:    CreateCoursePeriodRequest{},
			resp: CreateCoursePeriodResponse{
				ID: "1",
			},
		},
		{
			name: "error creating course period",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(-1), echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"))
			},
			userID:  &userID,
			req:     CreateCoursePeriodRequest{},
			wantErr: echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error"),
		},
		{
			name:    "error cannot find userID in context",
			svc:     &mocks.ServiceMock{},
			req:     CreateCoursePeriodRequest{},
			wantErr: echo.NewHTTPError(http.StatusBadRequest, "Bad Request"),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.ServiceMock{},
			req:     "bad request",
			wantErr: echo.NewHTTPError(http.StatusBadRequest, "Bad Request"),
		},
	}

	loggerMock := zap.NewNop()
	path := "/courses/:course_id/course_periods"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := MakeCoursePeriodEndpointsHandler(tt.svc, loggerMock)
			err = h.CreateCoursePeriod(ctx)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody CreateCoursePeriodResponse
				err = json.Unmarshal(rec.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestCoursePeriodEndpointsHandler_UpdateCoursePeriod(t *testing.T) {
	var userID = "1"
	type testCase struct {
		name    string
		svc     *mocks.ServiceMock
		prepare func(ctx echo.Context, tc *testCase)
		userID  *string
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("UpdateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).
					Return(nil)
			},
			userID: &userID,
			req:    UpdateCoursePeriodRequest{},
			resp:   UpdateCoursePeriodResponse{},
		},
		{
			name: "error updating course period",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("UpdateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).
					Return(errors.New("cannot update course period"))
			},
			userID:  &userID,
			req:     UpdateCoursePeriodRequest{},
			wantErr: echo.NewHTTPError(http.StatusUnprocessableEntity, "Unprocessable Entity"),
		},
		{
			name:    "error cannot find userID in context",
			svc:     &mocks.ServiceMock{},
			req:     UpdateCoursePeriodRequest{},
			wantErr: echo.NewHTTPError(http.StatusBadRequest, "Bad Request"),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.ServiceMock{},
			req:     "bad request",
			wantErr: echo.NewHTTPError(http.StatusBadRequest, "Bad Request"),
		},
	}

	loggerMock := zap.NewNop()
	path := "/courses/:course_id/course_periods/:id"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := MakeCoursePeriodEndpointsHandler(tt.svc, loggerMock)
			err = h.UpdateCoursePeriod(ctx)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody UpdateCoursePeriodResponse
				err = json.Unmarshal(rec.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestCoursePeriodEndpointsHandler_DeleteCoursePeriod(t *testing.T) {
	var userID = "1"
	type testCase struct {
		name    string
		svc     *mocks.ServiceMock
		prepare func(ctx echo.Context, tc *testCase)
		userID  *string
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("DeleteCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(nil)
			},
			userID: &userID,
			req:    DeleteCoursePeriodRequest{},
			resp:   DeleteCoursePeriodResponse{},
		},
		{
			name: "error deleting course period",
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("DeleteCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(errors.New("cannot delete course period"))
			},
			userID:  &userID,
			req:     DeleteCoursePeriodRequest{},
			wantErr: echo.NewHTTPError(http.StatusUnprocessableEntity, "Unprocessable Entity"),
		},
		{
			name:    "error cannot find userID in context",
			svc:     &mocks.ServiceMock{},
			req:     DeleteCoursePeriodRequest{},
			wantErr: echo.NewHTTPError(http.StatusBadRequest, "Bad Request"),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.ServiceMock{},
			req:     "bad request",
			wantErr: echo.NewHTTPError(http.StatusBadRequest, "Bad Request"),
		},
	}

	loggerMock := zap.NewNop()
	path := "/courses/:course_id/course_periods/:id"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := MakeCoursePeriodEndpointsHandler(tt.svc, loggerMock)
			err = h.DeleteCoursePeriod(ctx)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody DeleteCoursePeriodResponse
				err = json.Unmarshal(rec.Body.Bytes(), &respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}
