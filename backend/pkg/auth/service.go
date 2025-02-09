package auth

import (
	"context"
	"fmt"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/jwt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	GetUserRoles(ctx context.Context, userID int) ([]string, error)
}

// /generate service `AuthService` with `repository` field

func NewAuthService(repo Repository, signer jwt.Signer, logger log.Logger) Service {
	return &AuthService{
		repository: repo,
		signer:     signer,
		logger:     logger,
	}
}

type AuthService struct {
	repository Repository
	signer     jwt.Signer
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

	tokenString, err := s.signer.GenerateJWT(fmt.Sprintf("%d", user.ID), roles)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}
