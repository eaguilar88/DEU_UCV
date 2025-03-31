package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/pkg/auth/mocks"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMakeAuthEndpointsHandler(t *testing.T) {
	type testCase struct {
		name string
		svc  Service
		log  *zap.Logger
		want AuthEndpointsHandler
	}
	tc := testCase{
		name: "success",
		svc:  &mocks.ServiceMock{},
		log:  zap.NewNop(),
		want: AuthEndpointsHandler{
			svc: &mocks.ServiceMock{},
			log: zap.NewNop(),
		},
	}
	t.Run(tc.name, func(t *testing.T) {
		got := MakeAuthEndpointsHandler(tc.svc, tc.log)
		assert.Equal(t, tc.want.svc, got.svc)
	})
}

func TestAuthEndpointsHandler_LoginHandleHTTP(t *testing.T) {
	type testCase struct {
		name    string
		svc     *mocks.ServiceMock
		prepare func(ctx echo.Context, tc *testCase)
		req     any
		token   string
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			svc:  &mocks.ServiceMock{},
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
			svc:  &mocks.ServiceMock{},
			prepare: func(ctx echo.Context, tc *testCase) {
				tc.svc.On("Login", ctx.Request().Context(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tc.token, nil, tc.wantErr)
			},
			req:     LoginRequest{},
			token:   "",
			wantErr: echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized"),
		},
		{
			name:    "error cannot bind request",
			svc:     &mocks.ServiceMock{},
			req:     "invalid request",
			token:   "",
			wantErr: echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized"),
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

			h := MakeAuthEndpointsHandler(tt.svc, loggerMock)

			err = h.LoginHandleHTTP(ctx)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				var respBody LoginResponse
				err = json.Unmarshal(rec.Body.Bytes(), &respBody)
				assert.Equal(t, tt.token, respBody.Token) // Adjust as needed
				assert.Equal(t, tt.wantErr, err)
			}
			tt.svc.AssertExpectations(t)
		})
	}
}
