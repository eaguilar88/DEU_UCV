package users

import (
	"context"
	"slices"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Repository interface {
	GetUser(ctx context.Context, userID int) (entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int64, error)
	UpdateUser(ctx context.Context, userID int, user entities.User) error
	DeleteUser(ctx context.Context, userID int) error
	AddRoleToUser(ctx context.Context, userID, role int) error
}

type UserService struct {
	repo Repository
	log  log.Logger
}

func NewUsersService(repository Repository, logger log.Logger) Service {
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
		return -1, err
	}

	if slices.Contains(existingUser.Roles, user.Roles[0]) {
		err = s.repo.AddRoleToUser(ctx, existingUser.ID, entities.RoleIDFromName(user.Roles[0]))
		if err != nil {
			level.Error(s.log).Log("message", "error adding role to user", "error", err)
			return -1, err
		}
		return int64(user.ID), nil
	}

	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return -1, err
	}
	return id, nil
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
