package courses

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/courses/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// buildCourseForm builds a valid multipart form for CreateCourse with all required fields and a cover file.
func buildCourseForm() (*bytes.Buffer, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fields := map[string]string{
		"nombre":                "Test Course",
		"descripcion":           "A test description",
		"objetivos":             "Test objectives",
		"fundamentacion":        "Test rationale",
		"duracion":              "40 horas",
		"estructura_costos":     "1000 BsS",
		"perfil_docente":        "Instructor profile",
		"exigencias":            "Requirements",
		"estructura_curricular": "Course content",
		"evaluacion":            "Evaluation method",
		"cronograma":            "Schedule",
		"facultad":              "DEU",
	}
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	fw, _ := w.CreateFormFile("portada", "cover.jpg")
	_, _ = fw.Write([]byte("fake image data"))
	w.Close()

	return &b, w.FormDataContentType()
}

func TestNewHandler(t *testing.T) {
	svc := &mocks.MockService{}
	log := zap.NewNop()

	got := NewHandler(svc, log)
	assert.Equal(t, svc, got.svc)
	assert.Equal(t, log, got.log)
}

func TestHandler_GetCourse(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success with latest period",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCourse", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.Course{ID: "1", Name: "Test Course"}, nil)
				tc.svc.On("GetLatestCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.CoursePeriod{ID: "period-1", StartDate: "2026-01-01"}, nil)
			},
			resp: GetCourseResponse{
				ID:   "1",
				Name: "Test Course",
				LatestPeriod: &LatestCoursePeriodInfo{
					ID:        "period-1",
					StartDate: "2026-01-01",
				},
			},
		},
		{
			name: "success with no latest period",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCourse", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.Course{ID: "1", Name: "Test Course"}, nil)
				tc.svc.On("GetLatestCoursePeriod", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.CoursePeriod{}, errors.New("not found"))
			},
			resp: GetCourseResponse{
				ID:   "1",
				Name: "Test Course",
			},
		},
		{
			name: "error getting course",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCourse", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(entities.Course{}, errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/courses/1", nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("1")

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.GetCourse(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetCourseResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetCourses(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success empty list",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCourses", ctx.Request().Context(), mock.AnythingOfType("entities.PageScope")).
					Return([]entities.Course{}, entities.PageScope{}, nil)
			},
			resp: GetCoursesResponse{},
		},
		{
			name: "success with courses",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCourses", ctx.Request().Context(), mock.AnythingOfType("entities.PageScope")).
					Return([]entities.Course{{ID: "1", Name: "Course A"}}, entities.PageScope{Page: 1, PerPage: 10}, nil)
			},
			resp: GetCoursesResponse{
				Courses: []GetCourseResponse{{ID: "1", Name: "Course A"}},
				Pages:   entities.PageScope{Page: 1, PerPage: 10},
			},
		},
		{
			name: "error getting courses",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetCourses", ctx.Request().Context(), mock.AnythingOfType("entities.PageScope")).
					Return(nil, entities.PageScope{}, errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/courses", nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.GetCourses(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetCoursesResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetPublicCourses(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		resp    any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success empty list",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetPublicCourses", ctx.Request().Context(), mock.AnythingOfType("entities.PageScope")).
					Return([]entities.Course{}, entities.PageScope{}, nil)
			},
			resp: GetCoursesResponse{},
		},
		{
			name: "success with courses",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetPublicCourses", ctx.Request().Context(), mock.AnythingOfType("entities.PageScope")).
					Return([]entities.Course{{ID: "1", Name: "Course A"}}, entities.PageScope{Page: 1, PerPage: 10}, nil)
			},
			resp: GetCoursesResponse{
				Courses: []GetCourseResponse{{ID: "1", Name: "Course A"}},
				Pages:   entities.PageScope{Page: 1, PerPage: 10},
			},
		},
		{
			name: "error getting public courses",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("GetPublicCourses", ctx.Request().Context(), mock.AnythingOfType("entities.PageScope")).
					Return(nil, entities.PageScope{}, errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/courses/public", nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.GetPublicCourses(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody GetCoursesResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestHandler_CreateCourse(t *testing.T) {
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
			name:   "error missing required field",
			svc:    &mocks.MockService{},
			userID: &userID,
			buildForm: func() (*bytes.Buffer, string) {
				var b bytes.Buffer
				w := multipart.NewWriter(&b)
				w.Close()
				return &b, w.FormDataContentType()
			},
			wantErr: httperrors.NewBadRequest(NameMissingError),
		},
		{
			name:      "error from service",
			svc:       &mocks.MockService{},
			userID:    &userID,
			buildForm: buildCourseForm,
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCourse", ctx.Request().Context(), userID, mock.AnythingOfType("entities.Course")).
					Return(int64(-1), errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
		{
			name:      "success",
			svc:       &mocks.MockService{},
			userID:    &userID,
			buildForm: buildCourseForm,
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("CreateCourse", ctx.Request().Context(), userID, mock.AnythingOfType("entities.Course")).
					Return(int64(42), nil)
			},
			resp: CreateCoursesResponse{ID: "42"},
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

			req := httptest.NewRequest(http.MethodPost, "/courses", body)
			req.Header.Set(echo.HeaderContentType, contentType)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.CreateCourse(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody CreateCoursesResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
				assert.Equal(t, tt.resp, respBody)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateCourse(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		req     any
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("UpdateCourse", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.Course")).
					Return(nil)
			},
			req: UpdateCourseRequest{},
		},
		{
			name: "error from service",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("UpdateCourse", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("entities.Course")).
					Return(errors.New("db error"))
			},
			req:     UpdateCourseRequest{},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
		{
			name:    "error cannot bind",
			svc:     &mocks.MockService{},
			req:     "bad request",
			wantErr: httperrors.NewBadRequest("invalid request body"),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPatch, "/courses/1", bytes.NewBuffer(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("1")

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err = h.UpdateCourse(ctx)
			assertCustomError(t, tt.wantErr, err)
			tt.svc.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteCourse(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("DeleteCourse", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(nil)
			},
		},
		{
			name: "error from service",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("DeleteCourse", ctx.Request().Context(), mock.AnythingOfType("string")).
					Return(errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
	}

	logger := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/courses/1", nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("1")

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc, logger)
			err := h.DeleteCourse(ctx)
			assertCustomError(t, tt.wantErr, err)
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
