package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/auth/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	type testCase struct {
		name string
		svc  Service
		want Handler
	}
	tc := testCase{
		name: "success",
		svc:  &mocks.MockService{},
		want: Handler{
			svc: &mocks.MockService{},
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := NewHandler(tc.svc)
		assert.Equal(t, tc.want.svc, got.svc)
	})
}

func TestAuthEndpointsHandler_LoginHandleHTTP(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.MockService
		prepare func(ctx echo.Context, tc *testCase)
		req     any
		token   string
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("Login", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tc.token, &entities.User{}, tc.wantErr)
			},
			req:     LoginRequest{},
			token:   "token",
			wantErr: nil,
		},
		{
			name: "error login failed",
			svc:  &mocks.MockService{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("Login", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tc.token, nil, tc.wantErr)
			},
			req:     LoginRequest{},
			token:   "",
			wantErr: httperrors.NewUnauthorized("invalid credentials"),
		},
		{
			name:    "error cannot bind request",
			svc:     &mocks.MockService{},
			req:     "invalid request",
			token:   "",
			wantErr: httperrors.NewUnauthorized("invalid credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}

			h := NewHandler(tt.svc)

			err = h.LoginHandleHTTP(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody LoginResponse
				err = json.Unmarshal(rec.Body.Bytes(), &respBody)
				assert.Equal(t, tt.token, respBody.Token) // Adjust as needed
				require.NoError(t, err)
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

func TestAuthEndpointsHandler_Login_Roles(t *testing.T) {
	svc := mocks.NewMockService(t)
	body, err := json.Marshal(LoginRequest{Username: "proveedor@example.test", Password: "secreto"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)
	svc.EXPECT().Login(mock.Anything, "proveedor@example.test", "secreto").
		Return("token", &entities.User{ID: "15", Roles: []string{"course_admin"}}, nil)

	require.NoError(t, NewHandler(svc).LoginHandleHTTP(ctx))

	var resp LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	// roles keeps the backend names (grupos reads them); rol is the UI name.
	assert.Equal(t, []string{"course_admin"}, resp.User.Roles)
	assert.Equal(t, "proveedor", resp.User.Rol)
}

// newJSONContext builds an echo context with the app's validator for a JSON request.
func newJSONContext(t *testing.T, method, path string, body any) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	jsonBytes, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(method, path, bytes.NewReader(jsonBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	e.Validator = security.NewCustomValidator()
	return e.NewContext(req, rec), rec
}

func TestAuthEndpointsHandler_ChangePassword(t *testing.T) {
	validBody := ChangePasswordRequest{CurrentPassword: "nolodire", NewPassword: "nuevaClave123"}
	tests := []struct {
		name       string
		userID     string
		body       any
		prepare    func(svc *mocks.MockService)
		wantErr    error
		wantStatus int
	}{
		{
			name:   "success",
			userID: "1",
			body:   validBody,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ChangePassword(mock.Anything, "1", "nolodire", "nuevaClave123").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:    "no authenticated user",
			body:    validBody,
			wantErr: httperrors.NewUnauthorized("authentication required"),
		},
		{
			name:    "new password too short",
			userID:  "1",
			body:    ChangePasswordRequest{CurrentPassword: "nolodire", NewPassword: "corta"},
			wantErr: httperrors.NewBadRequest("validation failed"),
		},
		{
			name:   "wrong current password",
			userID: "1",
			body:   validBody,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ChangePassword(mock.Anything, "1", "nolodire", "nuevaClave123").Return(ErrInvalidCurrentPassword)
			},
			wantErr: httperrors.NewBadRequest("contraseña actual incorrecta"),
		},
		{
			name:   "service error",
			userID: "1",
			body:   validBody,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ChangePassword(mock.Anything, "1", "nolodire", "nuevaClave123").Return(errors.New("db error"))
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
			ctx, rec := newJSONContext(t, http.MethodPut, "/auth/password", tt.body)
			if tt.userID != "" {
				ctx.Set("userID", tt.userID)
			}

			err := NewHandler(svc).ChangePasswordHandleHTTP(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Equal(t, tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestAuthEndpointsHandler_ForgotPassword(t *testing.T) {
	tests := []struct {
		name    string
		body    any
		prepare func(svc *mocks.MockService)
		wantErr error
	}{
		{
			// The service returns nil for unknown emails too, so both answer the same 200.
			name: "success",
			body: ForgotPasswordRequest{Email: "ana.perez@example.test"},
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RequestPasswordReset(mock.Anything, "ana.perez@example.test").Return(nil)
			},
		},
		{
			name:    "invalid email",
			body:    ForgotPasswordRequest{Email: "no-es-un-correo"},
			wantErr: httperrors.NewBadRequest("validation failed"),
		},
		{
			name: "service error",
			body: ForgotPasswordRequest{Email: "ana.perez@example.test"},
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().RequestPasswordReset(mock.Anything, "ana.perez@example.test").Return(errors.New("db error"))
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
			ctx, rec := newJSONContext(t, http.MethodPost, "/auth/password/forgot", tt.body)

			err := NewHandler(svc).ForgotPasswordHandleHTTP(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Equal(t, http.StatusOK, rec.Code)
				var resp ForgotPasswordResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, forgotPasswordMessage, resp.Message)
			}
		})
	}
}

func TestAuthEndpointsHandler_ResetPassword(t *testing.T) {
	validBody := ResetPasswordRequest{Token: "reset-token", NewPassword: "nuevaClave123"}
	tests := []struct {
		name    string
		body    any
		prepare func(svc *mocks.MockService)
		wantErr error
	}{
		{
			name: "success",
			body: validBody,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ResetPassword(mock.Anything, "reset-token", "nuevaClave123").Return(nil)
			},
		},
		{
			name:    "missing token",
			body:    ResetPasswordRequest{NewPassword: "nuevaClave123"},
			wantErr: httperrors.NewBadRequest("validation failed"),
		},
		{
			name: "invalid or expired token",
			body: validBody,
			prepare: func(svc *mocks.MockService) {
				svc.EXPECT().ResetPassword(mock.Anything, "reset-token", "nuevaClave123").Return(ErrInvalidResetToken)
			},
			wantErr: httperrors.NewBadRequest("enlace inválido o expirado"),
		},
		{
			name: "service error",
			body: validBody,
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
			ctx, rec := newJSONContext(t, http.MethodPost, "/auth/password/reset", tt.body)

			err := NewHandler(svc).ResetPasswordHandleHTTP(ctx)
			assertCustomError(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Equal(t, http.StatusNoContent, rec.Code)
			}
		})
	}
}
