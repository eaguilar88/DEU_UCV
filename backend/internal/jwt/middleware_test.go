package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
