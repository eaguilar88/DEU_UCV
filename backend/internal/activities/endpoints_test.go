package activities

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/activities/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/security"
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

func TestHandler_GetActivities(t *testing.T) {
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
			name: "group_id only, no dates",
			url:  "/activities?group_id=1&page=1&per_page=10",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetActivities", ctx.Request().Context(), entities.ActivityFilter{GroupID: "1"}, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.Activity{{ID: "1"}}, entities.PageScope{Page: 1, PerPage: 10}, entities.ActivityMetrics{}, nil)
			},
			resp: GetActivitiesResponse{
				Activities: []GetActivityResponse{{ID: "1", KnowledgeArea: make([]string, 0)}},
				PageScope:  entities.PageScope{Page: 1, PerPage: 10},
			},
		},
		{
			name: "start and end date filter passed through",
			url:  "/activities?group_id=1&page=1&per_page=10&start_date=01-01-2026&end_date=31-01-2026",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetActivities", ctx.Request().Context(), entities.ActivityFilter{GroupID: "1", StartDate: "2026-01-01", EndDate: "2026-01-31"}, mock.AnythingOfType("entities.PageScope")).
					Return([]entities.Activity{{ID: "1"}}, entities.PageScope{}, entities.ActivityMetrics{}, nil)
			},
			resp: GetActivitiesResponse{
				Activities: []GetActivityResponse{{ID: "1", KnowledgeArea: make([]string, 0)}},
			},
		},
		{
			name:    "end date without start date returns bad request",
			url:     "/activities?group_id=1&page=1&per_page=10&end_date=31-01-2026",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewBadRequest("start_date is required when end_date is provided"),
		},
		{
			name:    "start date after end date returns bad request",
			url:     "/activities?group_id=1&page=1&per_page=10&start_date=31-01-2026&end_date=01-01-2026",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewBadRequest("start_date must not be after end_date"),
		},
		{
			name:    "malformed start date returns bad request",
			url:     "/activities?group_id=1&page=1&per_page=10&start_date=2026-01-01",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewBadRequest("start_date must be in DD-MM-YYYY format"),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			e.Validator = security.NewCustomValidator()
			ctx := e.NewContext(req, rec)

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.GetActivities(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetActivitiesResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
		})
	}
}

// buildActivityForm builds a valid multipart form for CreateActivity with the required
// fields and a cover image file.
func buildActivityForm() (*bytes.Buffer, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fields := map[string]string{
		"group_id":    "1",
		"nombre":      "Workshop",
		"descripcion": "A test activity",
		"gallery_url": "https://example.com/gallery",
	}
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	fw, _ := w.CreateFormFile(entities.ActivityFileTypeCoverImage, "cover.jpg")
	_, _ = fw.Write([]byte("fake image data"))
	w.Close()

	return &b, w.FormDataContentType()
}

func TestHandler_CreateActivity(t *testing.T) {
	userID := "user-1"

	type testCase struct {
		name      string
		svc       *mocks.MockService
		prepare   func(ctx echo.Context, tc *testCase)
		userID    *string
		buildForm func() (*bytes.Buffer, string)
		resp      any
		wantErr   error
	}

	tests := []testCase{
		{
			name:    "error missing userID",
			svc:     &mocks.MockService{},
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
		{
			name:      "error from service",
			svc:       &mocks.MockService{},
			userID:    &userID,
			buildForm: buildActivityForm,
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateActivity", ctx.Request().Context(), mock.AnythingOfType("entities.Activity")).
					Return(int64(-1), errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
		{
			name:      "success",
			svc:       &mocks.MockService{},
			userID:    &userID,
			buildForm: buildActivityForm,
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateActivity", ctx.Request().Context(), mock.AnythingOfType("entities.Activity")).
					Return(int64(42), nil)
			},
			resp: CreateActivityResponse{ID: "42"},
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			contentType := echo.MIMEApplicationJSON

			if tt.buildForm != nil {
				b, ct := tt.buildForm()
				body = b
				contentType = ct
			} else {
				body = &bytes.Buffer{}
			}

			req := httptest.NewRequest(http.MethodPost, "/activities", body)
			req.Header.Set(echo.HeaderContentType, contentType)
			rec := httptest.NewRecorder()
			e := echo.New()
			e.Validator = security.NewCustomValidator()
			ctx := e.NewContext(req, rec)

			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.CreateActivity(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody CreateActivityResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestHandler_buildDateFilter_StartDateOnlyDefaultsEndToToday(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/activities?start_date=01-01-2020", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	h := NewHandler(&mocks.MockService{}, zap.NewNop())
	start, end, err := h.buildDateFilter(ctx)

	require.NoError(t, err)
	assert.Equal(t, "2020-01-01", start)
	assert.Equal(t, time.Now().Format(storageDateFormat), end)
}
