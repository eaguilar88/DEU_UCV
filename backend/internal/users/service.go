package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetUser(ctx context.Context, userID string) (*entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int64, error)
	UpdateUser(ctx context.Context, userID string, user entities.User) error
	DeleteUser(ctx context.Context, userID string) error
	AddRoleToUser(ctx context.Context, tx *sql.Tx, userID string, role int) error
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

type MailClient interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

type UserService struct {
	repo        Repository
	emailClient MailClient
	log         *zap.Logger
}

func NewUsersService(repository Repository, emailClient MailClient, logger *zap.Logger) *UserService {
	return &UserService{
		repo:        repository,
		emailClient: emailClient,
		log:         logger,
	}
}

func (s *UserService) GetUser(ctx context.Context, userID string) (entities.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return entities.User{}, err
	}
	return *user, nil
}

func (s *UserService) GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error) {
	users, page, err := s.repo.GetUsers(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return users, page, nil
}

func (s *UserService) CreateUser(ctx context.Context, user entities.User) (int64, error) {
	existingUser, err := s.repo.GetUserByUsername(ctx, user.Username)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return -1, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil && existingUser.ID != "" {
		return -1, ErrUserAlreadyExists
	}

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return -1, fmt.Errorf("error creating new user: %w", err)
	}

	if err := s.emailClient.Send(ctx, user.Username, "Welcome to DEU", "Welcome to DEU"); err != nil {
		return -1, fmt.Errorf("error sending welcome email: %w", err)
	}
	return id, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, user entities.User) error {
	userFromDB, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	user.CI = userFromDB.CI
	user.Username = userFromDB.Username
	user.Password = userFromDB.Password

	if err := s.repo.UpdateUser(ctx, userID, user); err != nil {
		return err
	}
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}
	return nil
}
