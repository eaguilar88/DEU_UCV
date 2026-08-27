package course_periods

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eaguilar88/deu/internal/course_periods/mocks"
	"github.com/eaguilar88/deu/internal/courses"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var e *echo.Echo

func TestMain(m *testing.M) {
	e = echo.New()
	e.Validator = security.NewCustomValidator()
	os.Exit(m.Run())
}

func TestNewHandler(t *testing.T) {
	type testCase struct {
		name string
		svc  Service
		log  *zap.Logger
		want Handler
	}
	s := mocks.NewMockService(t)
	tc := testCase{
		name: "success",
		svc:  s,
		log:  zap.NewNop(),
		want: Handler{
			svc: s,
			log: zap.NewNop(),
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := NewHandler(tc.svc, tc.log)
		assert.Equal(t, tc.want.svc, got.svc)
	})
}

func TestHandler_GetCoursePeriod(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
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
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.CoursePeriod{}, errors.New("cannot get course period"))
			},
			req:     GetCoursePeriodRequest{},
			resp:    GetCoursePeriodResponse{},
			wantErr: httperrors.NewInternal(errors.New("cannot get course period")),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.MockService{},
			req:     "bad request",
			wantErr: httperrors.NewBadRequest("invalid request"),
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
			ctx := e.NewContext(req, rec)
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := NewHandler(tt.svc, loggerMock)
			err = h.GetCoursePeriod(ctx)
			assertCustomError(t, tt.wantErr, err)
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

func TestHandler_GetCoursePeriods(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		req     any
		resp    any
		wantErr error
	}
	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
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
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCoursePeriods", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.PageScope")).
					Return([]entities.CoursePeriod{}, entities.PageScope{}, errors.New("internal error"))
			},
			req:     GetCoursePeriodRequest{},
			resp:    GetCoursePeriodResponse{},
			wantErr: httperrors.NewInternal(errors.New("internal error")),
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
			ctx := e.NewContext(req, rec)
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := NewHandler(tt.svc, loggerMock)
			err = h.GetCoursePeriods(ctx)
			assertCustomError(t, tt.wantErr, err)
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

func TestHandler_CreateCoursePeriod(t *testing.T) {
	userID := "1"
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		userID  *string
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(1), nil)
			},
			userID: &userID,
			req:    CreateCoursePeriodRequest{Capacity: 30},
			resp: CreateCoursePeriodResponse{
				ID: "1",
			},
		},
		{
			name: "error creating course period",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(-1), errors.New("internal error"))
			},
			userID:  &userID,
			req:     CreateCoursePeriodRequest{Capacity: 30},
			wantErr: httperrors.NewInternal(errors.New("internal error")),
		},
		{
			name: "error course not found or not approved",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(-1), courses.ErrCourseNotFound)
			},
			userID:  &userID,
			req:     CreateCoursePeriodRequest{Capacity: 30},
			wantErr: httperrors.NewNotFound("course not found"),
		},
		{
			name: "error a course period is already open",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(-1), ErrCoursePeriodAlreadyOpen)
			},
			userID:  &userID,
			req:     CreateCoursePeriodRequest{Capacity: 30},
			wantErr: httperrors.NewConflict(ErrCoursePeriodAlreadyOpen.Error()),
		},
		{
			name: "error course has a pending closure request",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(-1), courses.ErrCourseClosureRequestPending)
			},
			userID:  &userID,
			req:     CreateCoursePeriodRequest{Capacity: 30},
			wantErr: httperrors.NewConflict(courses.ErrCourseClosureRequestPending.Error()),
		},
		{
			name:    "error cannot find userID in context",
			svc:     &mocks.MockService{},
			req:     CreateCoursePeriodRequest{Capacity: 30},
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
		{
			name: "error invalid capacity",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("entities.CoursePeriod")).
					Return(int64(-1), ErrInvalidCapacity)
			},
			userID:  &userID,
			req:     CreateCoursePeriodRequest{},
			wantErr: httperrors.NewBadRequest(ErrInvalidCapacity.Error()),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.MockService{},
			req:     "bad request",
			wantErr: httperrors.NewBadRequest("invalid request body"),
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
			ctx := e.NewContext(req, rec)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := NewHandler(tt.svc, loggerMock)
			err = h.CreateCoursePeriod(ctx)
			assertCustomError(t, tt.wantErr, err)
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

func TestHandler_UpdateCoursePeriod(t *testing.T) {
	userID := "1"
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		userID  *string
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("UpdateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).
					Return(nil)
			},
			userID: &userID,
			req:    UpdateCoursePeriodRequest{Capacity: 30},
			resp:   UpdateCoursePeriodResponse{},
		},
		{
			name: "error updating course period",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("UpdateCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).
					Return(errors.New("cannot update course period"))
			},
			userID:  &userID,
			req:     UpdateCoursePeriodRequest{Capacity: 30},
			wantErr: httperrors.NewInternal(errors.New("cannot update course period")),
		},
		{
			name:    "error cannot find userID in context",
			svc:     &mocks.MockService{},
			req:     UpdateCoursePeriodRequest{Capacity: 30},
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.MockService{},
			req:     "bad request",
			wantErr: httperrors.NewBadRequest("invalid request body"),
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
			ctx := e.NewContext(req, rec)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := NewHandler(tt.svc, loggerMock)
			err = h.UpdateCoursePeriod(ctx)
			assertCustomError(t, tt.wantErr, err)
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

func TestHandler_DeleteCoursePeriod(t *testing.T) {
	userID := "1"
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		userID  *string
		req     any
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
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
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("DeleteCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(errors.New("cannot delete course period"))
			},
			userID:  &userID,
			req:     DeleteCoursePeriodRequest{},
			wantErr: httperrors.NewInternal(errors.New("cannot delete course period")),
		},
		{
			name:    "error cannot find userID in context",
			svc:     &mocks.MockService{},
			req:     DeleteCoursePeriodRequest{},
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.MockService{},
			req:     "bad request",
			wantErr: httperrors.NewBadRequest("invalid request body"),
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
			ctx := e.NewContext(req, rec)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			h := NewHandler(tt.svc, loggerMock)
			err = h.DeleteCoursePeriod(ctx)
			assertCustomError(t, tt.wantErr, err)
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

// assertCustomError compares errors by status code and safe message
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
