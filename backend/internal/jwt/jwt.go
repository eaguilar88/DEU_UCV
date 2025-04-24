package jwt

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const tokenIssuer = "deu"

// CustomClaims defines the structure of the JWT claims
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// Signer defines the interface for an authorization service.
type Signer interface {
	ValidateToken(tokenString string) (map[string]any, error) // Takes token string
	GenerateJWT(userID string, roles []string) (string, error)
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

func (s *JWTSigner) GenerateJWT(userID string, roles []string) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": tokenIssuer,
		"exp": time.Now().
			Add(time.Second * time.Duration(s.TTL)).
			Unix(),
		// Set expiration using s.TTL
		"iat": time.Now().Unix(),
		"v1": map[string]interface{}{
			"roles":  roles,
			"userID": userID,
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
