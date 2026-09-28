package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRequireRoles(t *testing.T) {
	type testCase struct {
		name       string
		roles      []string
		setRoles   bool
		wantStatus int
		wantCalled bool
	}
	testCases := []testCase{
		{
			name:       "no roles in context",
			setRoles:   false,
			wantStatus: http.StatusForbidden,
			wantCalled: false,
		},
		{
			name:       "roles present but none match",
			roles:      []string{"visitante", "participante"},
			setRoles:   true,
			wantStatus: http.StatusForbidden,
			wantCalled: false,
		},
		{
			name:       "matching role allows request",
			roles:      []string{"visitante", "deu_admin"},
			setRoles:   true,
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/admin/anything", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tc.setRoles {
				c.Set("roles", tc.roles)
			}

			called := false
			next := func(c echo.Context) error {
				called = true
				return c.NoContent(http.StatusOK)
			}

			err := RequireRoles("root", "deu_admin", "faculty_admin")(next)(c)

			assert.Equal(t, tc.wantCalled, called)
			if tc.wantStatus == http.StatusForbidden {
				require.Error(t, err)
				var customErr httperrors.CustomError
				require.ErrorAs(t, err, &customErr)
				assert.Equal(t, http.StatusForbidden, customErr.StatusCode())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantStatus, rec.Code)
			}
		})
	}
}

func TestOptionalJWTMiddleware(t *testing.T) {
	signer := NewJWTSigner("test-signing-key", 3600, zap.NewNop())
	validToken, err := signer.GenerateJWT("42",
		[]entities.UserRole{{Name: "faculty_admin", DomainType: "faculty", Faculty: "FACES"}},
		"", "", "", "", "")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		authHeader  string
		wantUserID  string
		wantRoles   []string
		wantFaculty string
	}{
		{
			name: "no header continues anonymously",
		},
		{
			name:       "invalid token continues anonymously",
			authHeader: "Bearer not-a-jwt",
		},
		{
			name:        "valid token sets the claims",
			authHeader:  "Bearer " + validToken,
			wantUserID:  "42",
			wantRoles:   []string{"faculty_admin"},
			wantFaculty: "FACES",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/groups/1", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			called := false
			next := func(c echo.Context) error {
				called = true
				return c.NoContent(http.StatusOK)
			}

			err := OptionalJWTMiddleware(signer, zap.NewNop())(next)(c)

			require.NoError(t, err)
			assert.True(t, called, "the request must never be rejected")
			userID, _ := c.Get("userID").(string)
			roles, _ := c.Get("roles").([]string)
			faculty, _ := c.Get("faculty").(string)
			assert.Equal(t, tc.wantUserID, userID)
			assert.Equal(t, tc.wantRoles, roles)
			assert.Equal(t, tc.wantFaculty, faculty)
		})
	}
}
