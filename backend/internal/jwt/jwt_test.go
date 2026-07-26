package jwt

import (
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestJWTSigner_GenerateJWT_Faculty(t *testing.T) {
	type testCase struct {
		name        string
		roles       []entities.UserRole
		wantFaculty string
	}
	testCases := []testCase{
		{
			name:        "deu_admin role carries faculty DEU",
			roles:       []entities.UserRole{{Name: "deu_admin", DomainType: "all", Faculty: "DEU"}},
			wantFaculty: "DEU",
		},
		{
			name:        "faculty_admin role carries its own faculty",
			roles:       []entities.UserRole{{Name: "faculty_admin", DomainType: "all", Faculty: "Ciencias"}},
			wantFaculty: "Ciencias",
		},
		{
			name:        "role with no faculty leaves claim empty",
			roles:       []entities.UserRole{{Name: "visitante", DomainType: "all", Faculty: ""}},
			wantFaculty: "",
		},
	}

	signer := NewJWTSigner("test-signing-key", 3600, zap.NewNop())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenString, err := signer.GenerateJWT("user-id", tc.roles, "")
			require.NoError(t, err)

			claims, err := signer.ValidateToken(tokenString)
			require.NoError(t, err)

			v1Claims, ok := claims["v1"].(map[string]any)
			require.True(t, ok)

			assert.Equal(t, tc.wantFaculty, v1Claims["faculty"])
		})
	}
}
