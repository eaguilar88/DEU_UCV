package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
	"github.com/golang-jwt/jwt/v4"
)

const tokenIssuer = "deu"

type Repository interface {
	ValidateUser(ctx context.Context, username, password string) (entities.User, error)
}

// /generate service `AuthService` with `repository` field

func NewAuthService(key string, ttl uint32, repo Repository, logger log.Logger) *AuthService {
	return &AuthService{
		SigningKey: key,
		TTL:        ttl,
		repository: repo,
		logger:     logger,
	}
}

type AuthService struct {
	SigningKey string
	TTL        uint32
	repository Repository
	logger     log.Logger
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {

	user, err := s.repository.ValidateUser(ctx, username, password)
	if err != nil {
		return "", err // Return an error if validation fails
	}

	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": tokenIssuer,
		"exp": time.Now().Add(time.Second * time.Duration(s.TTL)).Unix(), // Set expiration using s.TTL
		"iat": time.Now().Unix(),
		"v1": map[string]interface{}{
			"role":   user.Role.ID,
			"userID": user.ID,
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
