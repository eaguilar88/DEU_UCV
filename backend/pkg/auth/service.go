package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const tokenIssuer = "deu"

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	GetUserRoles(ctx context.Context, userID int) ([]string, error)
}

// /generate service `AuthService` with `repository` field

func NewAuthService(key string, ttl uint32, repo Repository, logger log.Logger) Service {
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

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *entities.User, error) {

	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		level.Error(s.logger).Log("message", "invalid password", "error", err)
		return "", nil, err
	}

	roles, err := s.repository.GetUserRoles(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}

	user.Roles = roles

	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": tokenIssuer,
		"exp": time.Now().Add(time.Second * time.Duration(s.TTL)).Unix(), // Set expiration using s.TTL
		"iat": time.Now().Unix(),
		"v1": map[string]interface{}{
			"roles":  roles,
			"userID": user.ID,
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(s.SigningKey))
	if err != nil {
		return "", nil, fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, &user, nil
}
