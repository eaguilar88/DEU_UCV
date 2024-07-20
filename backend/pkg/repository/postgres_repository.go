package repository

import (
	"context"
	"database/sql"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
	"github.com/go-kit/log"
	"github.com/lib/pq"
)

const (
	pgErrorCodeUniqueViolation = "23505"
)

type scannable interface {
	Scan(dest ...interface{}) error
}

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
		user, err = scanUser(rows)
		if err != nil {
			return entities.User{}, NewQueryError(errScan, err)
		}
	}

	return newUserFromModel(user), nil
}

func (r *PostgresRepository) GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error) {
	sql, args, err := getUsers(pageScope).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		usr, err := scanUser(rows)
		if err != nil {
			return nil, entities.PageScope{}, NewQueryError(errScan, err)
		}
		users = append(users, newUserFromModel(usr))
	}
	pageScope.Count = len(users)
	return users, pageScope, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user entities.User) (int64, error) {
	sql, args, err := insertUser(user).ToSql()
	if err != nil {
		return -1, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			return 0, NewQueryError(errUniqueIndexViolation, err)
		}
		return 0, NewQueryError(errBadLastInsertID, err)
	}

	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, userID int, user entities.User) error {
	sql, args, err := updateUserInfo(user, userID).ToSql()
	if err != nil {
		return NewQueryError(errBadQuery, err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return NewQueryError(errBadQuery, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errRowsNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, userID int) error {
	sql, args, err := deleteUser(userID).ToSql()
	if err != nil {
		return NewQueryError(errBadQuery, err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return NewQueryError(errBadQuery, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errRowsNotFound
	}

	return nil
}

func scanUser(row scannable) (models.User, error) {
	result := models.User{}
	err := row.Scan(
		&result.ID,
		&result.IDType,
		&result.CI,
		&result.Username,
		&result.FirstName,
		&result.LastName,
		&result.DateOfBirth,
		&result.Gender,
		&result.EducationLevel,
		&result.Address,
		&result.CreatedAt,
	)

	return result, err
}
