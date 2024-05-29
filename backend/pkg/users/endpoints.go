package users

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
)

type Service interface {
	GetUser(ctx context.Context, userID int, pageScope entities.PageScope) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.Pagination, error)
	CreateUser(ctx context.Context, pageScope entities.PageScope) (int, error)
	UpdateUser(ctx context.Context, pageScope entities.PageScope) error
	DeleteUser(ctx context.Context, pageScope entities.PageScope) error
}
