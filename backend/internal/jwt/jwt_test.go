package jwt

import (
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	gojwt "github.com/golang-jwt/jwt/v4"
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
			tokenString, err := signer.GenerateJWT("user-id", tc.roles, "", "", "", "", "")
			require.NoError(t, err)

			claims, err := signer.ValidateToken(tokenString)
			require.NoError(t, err)

			v1Claims, ok := claims["v1"].(map[string]any)
			require.True(t, ok)

			assert.Equal(t, tc.wantFaculty, v1Claims["faculty"])
		})
	}
}

func TestJWTSigner_GenerateJWT_GroupAndProviderClaims(t *testing.T) {
	signer := NewJWTSigner("test-signing-key", 3600, zap.NewNop())

	t.Run("all populated", func(t *testing.T) {
		tokenString, err := signer.GenerateJWT("user-id", nil, "ECP-generic", "group-1", "Grupo Uno", "ECP-abc123", "Proveedor Uno")
		require.NoError(t, err)

		claims, err := signer.ValidateToken(tokenString)
		require.NoError(t, err)

		v1Claims, ok := claims["v1"].(map[string]any)
		require.True(t, ok)

		assert.Equal(t, "group-1", v1Claims["groupID"])
		assert.Equal(t, "Grupo Uno", v1Claims["groupName"])
		assert.Equal(t, "ECP-abc123", v1Claims["providerID"])
		assert.Equal(t, "Proveedor Uno", v1Claims["providerName"])
	})

	t.Run("all empty", func(t *testing.T) {
		tokenString, err := signer.GenerateJWT("user-id", nil, "", "", "", "", "")
		require.NoError(t, err)

		claims, err := signer.ValidateToken(tokenString)
		require.NoError(t, err)

		v1Claims, ok := claims["v1"].(map[string]any)
		require.True(t, ok)

		assert.Equal(t, "", v1Claims["groupID"])
		assert.Equal(t, "", v1Claims["groupName"])
		assert.Equal(t, "", v1Claims["providerID"])
		assert.Equal(t, "", v1Claims["providerName"])
	})
}

func TestJWTSigner_GenerateJWT_UIRoles(t *testing.T) {
	signer := NewJWTSigner("test-signing-key", 3600, zap.NewNop())

	tokenString, err := signer.GenerateJWT("user-id", []entities.UserRole{{Name: "course_admin", DomainType: "course"}}, "", "", "", "", "")
	require.NoError(t, err)
	claims, err := signer.ValidateToken(tokenString)
	require.NoError(t, err)

	// The UI reads the top-level claims; authorization keeps reading v1.roles.
	assert.Equal(t, "proveedor", claims["rol"])
	assert.Equal(t, []any{"proveedor"}, claims["roles"])
	assert.Equal(t, []any{"course_admin"}, claims["v1"].(map[string]any)["roles"])
}

func TestJWTSigner_PasswordResetToken(t *testing.T) {
	const passwordHash = "$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa"
	signer := NewJWTSigner("test-signing-key", 3600, zap.NewNop())

	t.Run("round trip", func(t *testing.T) {
		token, err := signer.GeneratePasswordResetToken("7", passwordHash)
		require.NoError(t, err)

		userID, fingerprint, err := signer.ValidatePasswordResetToken(token)
		require.NoError(t, err)
		assert.Equal(t, "7", userID)
		assert.Equal(t, PasswordFingerprint(passwordHash), fingerprint)
		assert.NotContains(t, token, passwordHash)
	})

	t.Run("expires after the reset TTL", func(t *testing.T) {
		token, err := signer.GeneratePasswordResetToken("7", passwordHash)
		require.NoError(t, err)
		claims, err := signer.ValidateToken(token)
		require.NoError(t, err)

		exp := time.Unix(int64(claims["exp"].(float64)), 0)
		assert.WithinDuration(t, time.Now().Add(PasswordResetTTL), exp, 5*time.Second)
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		expired, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, gojwt.MapClaims{
			"sub":     "7",
			"purpose": passwordResetPurpose,
			"pwf":     PasswordFingerprint(passwordHash),
			"exp":     time.Now().Add(-time.Minute).Unix(),
		}).SignedString([]byte("test-signing-key"))
		require.NoError(t, err)

		_, _, err = signer.ValidatePasswordResetToken(expired)
		assert.ErrorIs(t, err, ErrInvalidPasswordResetToken)
	})

	t.Run("token signed with another key is rejected", func(t *testing.T) {
		other := NewJWTSigner("other-key", 3600, zap.NewNop())
		token, err := other.GeneratePasswordResetToken("7", passwordHash)
		require.NoError(t, err)

		_, _, err = signer.ValidatePasswordResetToken(token)
		assert.ErrorIs(t, err, ErrInvalidPasswordResetToken)
	})

	t.Run("session token is rejected", func(t *testing.T) {
		session, err := signer.GenerateJWT("7", nil, "", "", "", "", "")
		require.NoError(t, err)

		_, _, err = signer.ValidatePasswordResetToken(session)
		assert.ErrorIs(t, err, ErrInvalidPasswordResetToken)
	})
}

func TestPasswordFingerprint(t *testing.T) {
	a := PasswordFingerprint("hash-a")
	assert.Len(t, a, passwordFingerprintLen)
	assert.Equal(t, a, PasswordFingerprint("hash-a"))
	assert.NotEqual(t, a, PasswordFingerprint("hash-b"))
}
