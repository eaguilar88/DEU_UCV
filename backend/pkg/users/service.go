package users

import (
	"context"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
)

type Repository interface {
	GetUser(ctx context.Context, userID int) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int, error)
	UpdateUser(ctx context.Context, userID int, user entities.User) error
	DeleteUser(ctx context.Context, userID int) error
}

type UserService struct {
	repo Repository
	log  log.Logger
}

func NewUsersService(repository Repository, logger log.Logger) *UserService {
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

	return nil, entities.PageScope{}, nil
}
func (s *UserService) CreateUser(ctx context.Context, user entities.User) (int, error) {
	return -1, nil
}
func (s *UserService) UpdateUser(ctx context.Context, userID int, user entities.User) error {
	return nil
}
func (s *UserService) DeleteUser(ctx context.Context, userID int) error {
	return nil
}
