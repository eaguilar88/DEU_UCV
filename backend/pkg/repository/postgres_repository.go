package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
	"github.com/go-kit/log"
)

type PostgresRepository struct {
	db           *sql.DB
	documentsDir string
	logger       log.Logger
}

func NewDB(connection *sql.DB, directory string, logger log.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:           connection,
		documentsDir: directory,
		logger:       logger,
	}
}

func (r *PostgresRepository) GetUser(ctx context.Context, userID int, pageScope entities.PageScope) (entities.User, error) {
	query := fmt.Sprintf(getUserQuery, userCommon)
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.User{}, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, userID)
	if err = row.Err(); err != nil {
		return entities.User{}, err
	}
	var user models.User
	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Gender,
		&user.EducationLevel,
		&user.Address,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return newUserFromModel(user), err
	}
	return newUserFromModel(user), nil
}
func (r *PostgresRepository) GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.Pagination, error) {
	return nil, entities.Pagination{}, nil
}
func (r *PostgresRepository) CreateUser(ctx context.Context, pageScope entities.PageScope) (int, error) {
	return -1, nil
}
func (r *PostgresRepository) UpdateUser(ctx context.Context, pageScope entities.PageScope) error {
	return nil
}
func (r *PostgresRepository) DeleteUser(ctx context.Context, pageScope entities.PageScope) error {
	return nil
}
