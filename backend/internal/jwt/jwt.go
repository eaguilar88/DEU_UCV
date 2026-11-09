package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const tokenIssuer = "deu"

const (
	// passwordResetPurpose marks a token as a password reset link token. Session tokens don't carry it.
	passwordResetPurpose = "password_reset"
	// PasswordResetTTL is how long a password reset link is valid.
	PasswordResetTTL = 30 * time.Minute
	// passwordFingerprintLen is the number of hex characters of the password hash digest kept in the token.
	passwordFingerprintLen = 16
)

// ErrInvalidPasswordResetToken is returned when a password reset token is malformed, expired,
// badly signed or not a password reset token.
var ErrInvalidPasswordResetToken = errors.New("invalid password reset token")

// CustomClaims defines the structure of the JWT claims
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// Signer defines the interface for an authorization service.
type Signer interface {
	ValidateToken(tokenString string) (map[string]any, error) // Takes token string
	GenerateJWT(userID string, roles []entities.UserRole, providerCode, groupID, groupName, courseProviderID, courseProviderName string) (string, error)
	GeneratePasswordResetToken(userID, passwordHash string) (string, error)
	ValidatePasswordResetToken(tokenString string) (userID, fingerprint string, err error)
}

type JWTSigner struct {
	SigningKey string
	TTL        uint32
	Logger     *zap.Logger
}

func NewJWTSigner(signingKey string, ttl uint32, logger *zap.Logger) *JWTSigner {
	return &JWTSigner{
		SigningKey: signingKey,
		TTL:        ttl,
		Logger:     logger,
	}
}

func (s *JWTSigner) ValidateToken(tokenString string) (map[string]any, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.SigningKey), nil // Your signing key here
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Token is valid
	return claims, nil
}

func (s *JWTSigner) GenerateJWT(userID string, roles []entities.UserRole, providerCode, groupID, groupName, courseProviderID, courseProviderName string) (string, error) {
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	domainType := "all"
	faculty := ""
	for _, r := range roles {
		if r.DomainType != "all" {
			domainType = r.DomainType
		}
		if faculty == "" && r.Faculty != "" {
			faculty = r.Faculty
		}
	}

	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": tokenIssuer,
		"exp": time.Now().
			Add(time.Second * time.Duration(s.TTL)).
			Unix(),
		// Set expiration using s.TTL
		"iat": time.Now().Unix(),
		// The role names the diplomados UI compares against (e.g. "proveedor"). Authorization
		// only reads v1.roles.
		"rol":   entities.UIRole(roleNames),
		"roles": entities.UIRoles(roleNames),
		"v1": map[string]interface{}{
			"roles":        roleNames,
			"userID":       userID,
			"domainType":   domainType,
			"faculty":      faculty,
			"providerCode": providerCode,
			"groupID":      groupID,
			"groupName":    groupName,
			"providerID":   courseProviderID,
			"providerName": courseProviderName,
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(s.SigningKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, nil
}

// PasswordFingerprint derives a short digest of a password hash. A password reset token carries
// the fingerprint of the hash it was issued for, so the token stops working once the password changes.
func PasswordFingerprint(passwordHash string) string {
	sum := sha256.Sum256([]byte(passwordHash))
	return hex.EncodeToString(sum[:])[:passwordFingerprintLen]
}

// GeneratePasswordResetToken issues a short-lived token for the password reset link. It has no v1
// claims, so JWTMiddleware rejects it as a session token.
func (s *JWTSigner) GeneratePasswordResetToken(userID, passwordHash string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":     tokenIssuer,
		"sub":     userID,
		"purpose": passwordResetPurpose,
		"pwf":     PasswordFingerprint(passwordHash),
		"iat":     now.Unix(),
		"exp":     now.Add(PasswordResetTTL).Unix(),
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.SigningKey))
	if err != nil {
		return "", fmt.Errorf("error signing password reset token: %w", err)
	}
	return tokenString, nil
}

// ValidatePasswordResetToken checks a password reset token and returns the user ID and password
// fingerprint it was issued for. Session tokens are rejected.
func (s *JWTSigner) ValidatePasswordResetToken(tokenString string) (string, string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrInvalidPasswordResetToken, err)
	}

	if purpose, ok := claims["purpose"].(string); !ok || purpose != passwordResetPurpose {
		return "", "", fmt.Errorf("%w: wrong purpose", ErrInvalidPasswordResetToken)
	}
	userID, okUser := claims["sub"].(string)
	fingerprint, okFingerprint := claims["pwf"].(string)
	if !okUser || !okFingerprint || userID == "" || fingerprint == "" {
		return "", "", fmt.Errorf("%w: missing claims", ErrInvalidPasswordResetToken)
	}

	return userID, fingerprint, nil
}
