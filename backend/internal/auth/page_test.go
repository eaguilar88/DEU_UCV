package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/auth/mocks"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	pageInvalidText = "Enlace inválido o expirado"
	pageDoneText    = "Contraseña actualizada"
	pageFormText    = `name="confirmacion"`
	pageRedirect    = `<meta http-equiv="refresh" content="5;url=https://extension.ucv.ve/">`
	pageHomeLink    = `<a href="https://extension.ucv.ve/">`
)

// assertResetPageHeaders checks the headers that keep the token out of caches, Referer and frames.
func assertResetPageHeaders(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, "no-store", rec.Header().Get(echo.HeaderCacheControl))
	assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
	assert.Equal(t, "DENY", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Equal(t, resetPageCSP, rec.Header().Get(echo.HeaderContentSecurityPolicy))
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "text/html")
	assert.NotContains(t, rec.Body.String(), "<script")
}

func TestAuthEndpointsHandler_ResetPasswordPage(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		prepare    func(svc *mocks.MockService)
		wantErr    error
		wantStatus int
		wantText   string
	}{
		{
			name:  "valid token shows the form",
			query: "?token=reset-token",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().CheckPasswordResetToken(mock.Anything, "reset-token").Return(nil)
			},
			wantStatus: http.StatusOK,
			wantText:   `<input type="hidden" name="token" value="reset-token">`,
		},
		{
			name:       "missing token shows the invalid link page",
			wantStatus: http.StatusBadRequest,
			wantText:   pageInvalidText,
		},
		{
			name:  "invalid token shows the invalid link page",
			query: "?token=reset-token",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().CheckPasswordResetToken(mock.Anything, "reset-token").Return(ErrInvalidResetToken)
			},
			wantStatus: http.StatusBadRequest,
			wantText:   pageInvalidText,
		},
		{
			name:  "service error",
			query: "?token=reset-token",
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().CheckPasswordResetToken(mock.Anything, "reset-token").Return(errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.prepare != nil {
				tt.prepare(svc)
			}
			req := httptest.NewRequest(http.MethodGet, ResetPasswordPagePath+tt.query, nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			err := NewHandler(svc).ResetPasswordPage(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantStatus, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.wantText)
				assert.NotContains(t, rec.Body.String(), `http-equiv="refresh"`)
				assertResetPageHeaders(t, rec)
			}
		})
	}
}

func TestAuthEndpointsHandler_ResetPasswordPage_EscapesToken(t *testing.T) {
	const token = `"><script>alert(1)</script>`
	svc := mocks.NewMockService(t)
	svc.EXPECT().CheckPasswordResetToken(mock.Anything, token).Return(nil)
	req := httptest.NewRequest(http.MethodGet, ResetPasswordPagePath+"?token="+url.QueryEscape(token), nil)
	rec := httptest.NewRecorder()

	require.NoError(t, NewHandler(svc).ResetPasswordPage(echo.New().NewContext(req, rec)))

	assert.Equal(t, http.StatusOK, rec.Code)
	assertResetPageHeaders(t, rec)
}

func TestAuthEndpointsHandler_SubmitResetPasswordPage(t *testing.T) {
	tests := []struct {
		name       string
		form       url.Values
		prepare    func(svc *mocks.MockService)
		wantErr    error
		wantStatus int
		wantText   []string
	}{
		{
			name: "success",
			form: url.Values{"token": {"reset-token"}, "password": {"nuevaClave123"}, "confirmacion": {"nuevaClave123"}},
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ResetPassword(mock.Anything, "reset-token", "nuevaClave123").Return(nil)
			},
			wantStatus: http.StatusOK,
			wantText:   []string{pageDoneText, pageRedirect, pageHomeLink},
		},
		{
			name:       "passwords don't match",
			form:       url.Values{"token": {"reset-token"}, "password": {"nuevaClave123"}, "confirmacion": {"otraClave123"}},
			wantStatus: http.StatusBadRequest,
			wantText:   []string{"Las contraseñas no coinciden.", pageFormText, `value="reset-token"`},
		},
		{
			name:       "password too short",
			form:       url.Values{"token": {"reset-token"}, "password": {"corta"}, "confirmacion": {"corta"}},
			wantStatus: http.StatusBadRequest,
			wantText:   []string{"La contraseña debe tener al menos 8 caracteres.", pageFormText, `value="reset-token"`},
		},
		{
			name:       "missing token",
			form:       url.Values{"password": {"nuevaClave123"}, "confirmacion": {"nuevaClave123"}},
			wantStatus: http.StatusBadRequest,
			wantText:   []string{pageInvalidText},
		},
		{
			name: "invalid token",
			form: url.Values{"token": {"reset-token"}, "password": {"nuevaClave123"}, "confirmacion": {"nuevaClave123"}},
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ResetPassword(mock.Anything, "reset-token", "nuevaClave123").Return(ErrInvalidResetToken)
			},
			wantStatus: http.StatusBadRequest,
			wantText:   []string{pageInvalidText},
		},
		{
			name: "service error",
			form: url.Values{"token": {"reset-token"}, "password": {"nuevaClave123"}, "confirmacion": {"nuevaClave123"}},
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ResetPassword(mock.Anything, "reset-token", "nuevaClave123").Return(errors.New("db error"))
			},
			wantErr: httperrors.NewInternal(errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.prepare != nil {
				tt.prepare(svc)
			}
			req := httptest.NewRequest(http.MethodPost, ResetPasswordPagePath, strings.NewReader(tt.form.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			err := NewHandler(svc).SubmitResetPasswordPage(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantStatus, rec.Code)
				for _, text := range tt.wantText {
					assert.Contains(t, rec.Body.String(), text)
				}
				assertResetPageHeaders(t, rec)
			}
		})
	}
}
