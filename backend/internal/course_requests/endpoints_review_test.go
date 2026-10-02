package course_requests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/course_requests/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newReviewContext(body string, userID string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = security.NewCustomValidator()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("1")
	if userID != "" {
		ctx.Set("userID", userID)
	}
	return ctx, rec
}

func assertReviewResult(t *testing.T, err error, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	if wantStatus == http.StatusNoContent {
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String())
		return
	}
	var customErr httperrors.CustomError
	if assert.True(t, errors.As(err, &customErr), "expected CustomError, got %v", err) {
		assert.Equal(t, wantStatus, customErr.StatusCode())
	}
}

func TestHandler_RejectCourseRequest(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		userID     string
		prepare    func(svc *mocks.MockService)
		wantStatus int
	}{
		{
			name:   "rejects and returns 204",
			body:   `{"observaciones":"Documentación incompleta"}`,
			userID: "reviewer-1",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RejectCourseRequest(mock.Anything, "1", "reviewer-1", "Documentación incompleta").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing user returns 401",
			body:       `{}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "request not found returns 404",
			body:   `{}`,
			userID: "reviewer-1",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RejectCourseRequest(mock.Anything, "1", "reviewer-1", "").Return(ErrCourseRequestNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "service failure returns 500",
			body:   `{}`,
			userID: "reviewer-1",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RejectCourseRequest(mock.Anything, "1", "reviewer-1", "").Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.prepare != nil {
				tt.prepare(svc)
			}
			ctx, rec := newReviewContext(tt.body, tt.userID)

			err := NewHandler(svc, zap.NewNop()).RejectCourseRequest(ctx)

			assertReviewResult(t, err, rec, tt.wantStatus)
		})
	}
}

func TestHandler_RedirectCourseRequest(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		userID     string
		prepare    func(svc *mocks.MockService)
		wantStatus int
	}{
		{
			name:   "redirects and returns 204",
			body:   `{"facultad":"Ciencias","motivo":"Facultad incorrecta"}`,
			userID: "reviewer-1",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RedirectCourseRequest(mock.Anything, "1", "reviewer-1", entities.FacultyCiencias, "Facultad incorrecta").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing user returns 401",
			body:       `{"facultad":"Ciencias","motivo":"Facultad incorrecta"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing motivo returns 400",
			body:       `{"facultad":"Ciencias"}`,
			userID:     "reviewer-1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid faculty returns 400",
			body:       `{"facultad":"Inexistente","motivo":"Facultad incorrecta"}`,
			userID:     "reviewer-1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "request not found returns 404",
			body:   `{"facultad":"Ciencias","motivo":"Facultad incorrecta"}`,
			userID: "reviewer-1",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RedirectCourseRequest(mock.Anything, "1", "reviewer-1", entities.FacultyCiencias, "Facultad incorrecta").Return(ErrCourseRequestNotFound)
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
			ctx, rec := newReviewContext(tt.body, tt.userID)

			err := NewHandler(svc, zap.NewNop()).RedirectCourseRequest(ctx)

			assertReviewResult(t, err, rec, tt.wantStatus)
		})
	}
}
