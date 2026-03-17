package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/auth/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewHandler(t *testing.T) {
	type testCase struct {
		name string
		svc  Service
		log  *zap.Logger
		want Handler
	}
	tc := testCase{
		name: "success",
		svc:  &mocks.MockService{},
		log:  zap.NewNop(),
		want: Handler{
			svc: &mocks.MockService{},
			log: zap.NewNop(),
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := NewHandler(tc.svc, tc.log)
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
			wantErr: errors.NewUnauthorized("invalid credentials"),
		},
		{
			name:    "error cannot bind request",
			svc:     &mocks.MockService{},
			req:     "invalid request",
			token:   "",
			wantErr: errors.NewUnauthorized("invalid credentials"),
		},
	}

	loggerMock := zap.NewNop()
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

			h := NewHandler(tt.svc, loggerMock)

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
	expectedErr, ok := expected.(errors.CustomError)
	if !ok {
		t.Errorf("expected error is not CustomError: %T", expected)
		return
	}
	actualErr, ok := actual.(errors.CustomError)
	if !ok {
		t.Errorf("actual error is not CustomError: %T", actual)
		return
	}
	assert.Equal(t, expectedErr.StatusCode(), actualErr.StatusCode(), "status codes should match")
	assert.Equal(t, expectedErr.SafeMessage(), actualErr.SafeMessage(), "safe messages should match")
}
