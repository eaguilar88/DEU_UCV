package users

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
)

type Repository interface {
	GetUser(ctx context.Context, userID int, pageScope entities.PageScope) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.Pagination, error)
	CreateUser(ctx context.Context, user entities.User) (int, error)
	UpdateUser(ctx context.Context, user entities.User) error
	DeleteUser(ctx context.Context, userID int) error
}

type UserService struct {
	repo Repository
	log  log.Logger
}

func NewAppService(repository Repository, logger log.Logger) *UserService {
	return &UserService{
		repo: repository,
		log:  logger,
	}
}

func (s *UserService) GetUser(ctx context.Context, userID int, pageScope entities.PageScope) (entities.User, error) {
	return entities.User{}, nil
}
func (s *UserService) GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.Pagination, error) {

	return nil, entities.Pagination{}, nil
}
func (s *UserService) CreateUser(ctx context.Context, pageScope entities.PageScope) (int, error) {
	return -1, nil
}
func (s *UserService) UpdateUser(ctx context.Context, pageScope entities.PageScope) error {
	return nil
}
func (s *UserService) DeleteUser(ctx context.Context, pageScope entities.PageScope) error {
	return nil
}
