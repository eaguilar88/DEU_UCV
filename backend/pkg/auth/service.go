package auth

import (
	"context"

	"github.com/go-kit/log"
)

type Repository interface {
	Login(ctx context.Context, username, password string) (string, error)
}

// /generate service `AuthService` with `repository` field

func NewAuthService(repo Repository, logger log.Logger) *AuthService {
	return &AuthService{
		repository: repo,
		logger:     logger,
	}
}

type AuthService struct {
	repository Repository
	logger     log.Logger
}

// /generate service method `func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) { 	return "", nil }`
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	return "", nil
}

// /generate service method `func (s *AuthService) Register(ctx context.Context, username, password string) (string, error) { 	return "", nil }`
func (s *AuthService) Register(ctx context.Context, username, password string) (string, error) {
	return "", nil
}
