package repository

import (
	"context"
	"database/sql"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
	"github.com/go-kit/log"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db           *sql.DB
	documentsDir string
	logger       log.Logger
}

func NewRepository(connection *sql.DB, directory string, logger log.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:           connection,
		documentsDir: directory,
		logger:       logger,
	}
}

func (r *PostgresRepository) GetUser(ctx context.Context, userID int) (entities.User, error) {
	query := getUserByID(userID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.User{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		// level.Error(r.logger).Log("message", "failed to prepare", "error", err)
		return entities.User{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return entities.User{}, err
	}
	defer rows.Close()
	var user models.User
	for rows.Next() {
		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.DateOfBirth,
			&user.Gender,
			&user.EducationLevel,
			&user.Address,
		); err != nil {
			return entities.User{}, err
		}
	}

	return newUserFromModel(user), nil
}
func (r *PostgresRepository) GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error) {
	return nil, entities.PageScope{}, nil
}
func (r *PostgresRepository) CreateUser(ctx context.Context, user entities.User) (int, error) {
	return -1, nil
}
func (r *PostgresRepository) UpdateUser(ctx context.Context, userID int, user entities.User) error {
	return nil
}
func (r *PostgresRepository) DeleteUser(ctx context.Context, userID int) error {
	return nil
}
