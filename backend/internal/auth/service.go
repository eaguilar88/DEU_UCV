package auth

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

func NewAuthService(repo Repository, signer jwt.Signer, logger *zap.Logger) Service {
	return &AuthService{
		repository: repo,
		signer:     signer,
		logger:     logger,
	}
}

type AuthService struct {
	repository Repository
	signer     jwt.Signer
	logger     *zap.Logger
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *entities.User, error) {

	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, err
	}

	roles, err := s.repository.GetUserRoles(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}

	user.Roles = roles

	tokenString, err := s.signer.GenerateJWT(user.ID, roles)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}
