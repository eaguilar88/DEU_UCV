package auth

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]entities.UserRole, error)
	GetProviderCodeByUserID(ctx context.Context, userID string) (string, error)
}

type service struct {
	repository Repository
	signer     jwt.Signer
	logger     *zap.Logger
}

func NewService(repo Repository, signer jwt.Signer, logger *zap.Logger) Service {
	return &service{
		repository: repo,
		signer:     signer,
		logger:     logger,
	}
}

func (s *service) Login(ctx context.Context, username, password string) (string, *entities.User, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, err
	}

	userRoles, err := s.repository.GetUserRoles(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}

	roleNames := make([]string, len(userRoles))
	for i, r := range userRoles {
		roleNames[i] = r.Name
	}
	user.Roles = roleNames

	for _, r := range userRoles {
		if r.Faculty != "" {
			user.Faculty = r.Faculty
			break
		}
	}

	providerCode, err := s.repository.GetProviderCodeByUserID(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	user.ProviderCode = providerCode

	tokenString, err := s.signer.GenerateJWT(user.ID, userRoles, providerCode)
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}
