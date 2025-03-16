package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"slices"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	errs "github.com/eaguilar88/deu/pkg/errors"
	"go.uber.org/zap"
)

type Repository interface {
	GetUser(ctx context.Context, userID int) (entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int64, error)
	UpdateUser(ctx context.Context, userID int, user entities.User) error
	DeleteUser(ctx context.Context, userID int) error
	AddRoleToUser(ctx context.Context, tx *sql.Tx, userID, role int) error
	GetUserRoles(ctx context.Context, userID int) ([]string, error)
}

type UserService struct {
	repo Repository
	log  *zap.Logger
}

func NewUsersService(repository Repository, logger *zap.Logger) Service {
	return &UserService{
		repo: repository,
		log:  logger,
	}
}

func (s *UserService) GetUser(ctx context.Context, userID string) (entities.User, error) {
	intID, err := strconv.Atoi(userID)
	if err != nil {
		return entities.User{}, err
	}
	user, err := s.repo.GetUser(ctx, intID)
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
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
	if err != nil {
		if !errs.IsNotFoundError(err) {
			return -1, err
		}
		id, err := s.repo.CreateUser(ctx, user)
		if err != nil {
			return -1, err
		}

		return id, nil
	}

	// User exists, now handle roles
	roles, err := s.repo.GetUserRoles(ctx, existingUser.ID)
	if err != nil {
		s.log.Error("error getting user roles", zap.Error(err))
		return -1, err
	}

	if slices.Contains(roles, user.Roles[0]) {
		s.log.Warn(fmt.Sprintf("user %s already has the role %s", existingUser.Username, user.Roles[0]))
		return -1, errs.NewDuplicateEntryError(errors.New("user already exist"))
	}

	err = s.repo.AddRoleToUser(ctx, nil, existingUser.ID, entities.RoleIDFromName(user.Roles[0]))
	if err != nil {
		s.log.Error("error adding role to user", zap.Error(err))
		return -1, err
	}

	return int64(existingUser.ID), nil // Return the existing user's ID
}

func (s *UserService) UpdateUser(ctx context.Context, userID int, user entities.User) error {
	if err := s.repo.UpdateUser(ctx, userID, user); err != nil {
		return err
	}
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID int) error {
	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		return err
	}
	return nil
}
