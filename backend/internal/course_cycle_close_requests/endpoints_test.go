package course_cycle_close_requests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/course_cycle_close_requests/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	expectedCustom, ok := expected.(httperrors.CustomError)
	if !ok {
		assert.Equal(t, expected, actual)
		return
	}
	actualCustom, ok := actual.(httperrors.CustomError)
	if !ok {
		t.Errorf("expected a CustomError but got %T: %v", actual, actual)
		return
	}
	assert.Equal(t, expectedCustom.StatusCode(), actualCustom.StatusCode(), "status codes should match")
	assert.Equal(t, expectedCustom.SafeMessage(), actualCustom.SafeMessage(), "safe messages should match")
}

func TestHandler_SubmitCloseRequest(t *testing.T) {
	userID := "1"

	type testCase struct {
		name    string
		userID  *string
		prepare func(ctx echo.Context, svc *mocks.MockService)
		fields  map[string]string
		omit    []string
		wantErr error
	}

	tests := []testCase{
		{
			name:   "success",
			userID: &userID,
			prepare: func(ctx echo.Context, svc *mocks.MockService) {
				svc.EXPECT().SubmitCloseRequest(ctx.Request().Context(), mock.AnythingOfType("entities.CourseCycleCloseRequest"), userID).
					Return(int64(10), nil)
			},
			fields: map[string]string{"course_cycle_id": "5", "observaciones": "cierre"},
		},
		{
			name:    "error cannot find userID in context",
			fields:  map[string]string{"course_cycle_id": "5"},
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
		{
			name:    "error missing evidence file",
			userID:  &userID,
			fields:  map[string]string{"course_cycle_id": "5"},
			omit:    []string{entities.CloseRequestFileTypeParticipants},
			wantErr: httperrors.NewBadRequest("archivo_participantes is required"),
		},
		{
			name:   "error a close request is already pending",
			userID: &userID,
			prepare: func(ctx echo.Context, svc *mocks.MockService) {
				svc.EXPECT().SubmitCloseRequest(ctx.Request().Context(), mock.AnythingOfType("entities.CourseCycleCloseRequest"), userID).
					Return(int64(-1), ErrCloseRequestAlreadyPending)
			},
			fields:  map[string]string{"course_cycle_id": "5"},
			wantErr: httperrors.NewConflict(ErrCloseRequestAlreadyPending.Error()),
		},
		{
			name:   "error submitting close request",
			userID: &userID,
			prepare: func(ctx echo.Context, svc *mocks.MockService) {
				svc.EXPECT().SubmitCloseRequest(ctx.Request().Context(), mock.AnythingOfType("entities.CourseCycleCloseRequest"), userID).
					Return(int64(-1), errors.New("internal error"))
			},
			fields:  map[string]string{"course_cycle_id": "5"},
			wantErr: httperrors.NewInternal(errors.New("internal error")),
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			ctx := newMultipartCloseRequest(t, tt.fields, tt.omit...)
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, svc)
			}
			h := NewHandler(svc, loggerMock)
			err := h.SubmitCloseRequest(ctx)
			assertCustomError(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ApproveCloseRequest(t *testing.T) {
	userID := "1"
	type testCase struct {
		name    string
		userID  *string
		prepare func(ctx echo.Context, svc *mocks.MockService)
		wantErr error
	}

	tests := []testCase{
		{
			name:   "success",
			userID: &userID,
			prepare: func(ctx echo.Context, svc *mocks.MockService) {
				svc.EXPECT().ApproveCloseRequest(ctx.Request().Context(), "1", userID).Return(nil)
			},
		},
		{
			name:   "error not found",
			userID: &userID,
			prepare: func(ctx echo.Context, svc *mocks.MockService) {
				svc.EXPECT().ApproveCloseRequest(ctx.Request().Context(), "1", userID).Return(ErrCycleCloseRequestNotFound)
			},
			wantErr: httperrors.NewNotFound("close request not found"),
		},
		{
			name:    "error cannot find userID in context",
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			req := httptest.NewRequest(http.MethodPost, "/admin/course-cycle-close-requests/1/approve", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("1")
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, svc)
			}
			h := NewHandler(svc, loggerMock)
			err := h.ApproveCloseRequest(ctx)
			assertCustomError(t, tt.wantErr, err)
		})
	}
}

func TestHandler_RejectCloseRequest(t *testing.T) {
	userID := "1"
	type testCase struct {
		name    string
		userID  *string
		prepare func(ctx echo.Context, svc *mocks.MockService)
		wantErr error
	}

	tests := []testCase{
		{
			name:   "success",
			userID: &userID,
			prepare: func(ctx echo.Context, svc *mocks.MockService) {
				svc.EXPECT().RejectCloseRequest(ctx.Request().Context(), "1", userID, "").Return(nil)
			},
		},
		{
			name:    "error cannot find userID in context",
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			req := httptest.NewRequest(http.MethodPost, "/admin/course-cycle-close-requests/1/reject", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("1")
			if tt.userID != nil {
				ctx.Set("userID", *tt.userID)
			}
			if tt.prepare != nil {
				tt.prepare(ctx, svc)
			}
			h := NewHandler(svc, loggerMock)
			err := h.RejectCloseRequest(ctx)
			assertCustomError(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetCloseRequests(t *testing.T) {
	svc := mocks.NewMockService(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/course-cycle-close-requests", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	svc.EXPECT().GetCloseRequests(ctx.Request().Context(), entities.Faculty(""), mock.AnythingOfType("entities.PageScope")).
		Return([]entities.CourseCycleCloseRequest{{ID: 1}}, entities.PageScope{}, nil)

	h := NewHandler(svc, zap.NewNop())
	err := h.GetCloseRequests(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandler_GetCloseRequestByID(t *testing.T) {
	svc := mocks.NewMockService(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/course-cycle-close-requests/1", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("1")
	svc.EXPECT().GetCloseRequestByID(ctx.Request().Context(), "1").Return(entities.CourseCycleCloseRequest{ID: 1}, nil)

	h := NewHandler(svc, zap.NewNop())
	err := h.GetCloseRequestByID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
