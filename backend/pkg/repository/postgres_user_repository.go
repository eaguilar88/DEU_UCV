package repository

import (
	"context"
	"database/sql"

	"github.com/eaguilar88/deu/pkg/entities"
	errs "github.com/eaguilar88/deu/pkg/errors"
	"github.com/eaguilar88/deu/pkg/repository/models"
	"github.com/eaguilar88/deu/pkg/repository/queries"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/lib/pq"
)

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (entities.User, error) {
	query, args, err := queries.GetUserByUsername(username).ToSql()
	if err != nil {
		return entities.User{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.User{}, err
	}
	defer stmt.Close()

	var user models.User
	err = stmt.QueryRowContext(ctx, args...).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Password,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.User{}, errs.NewNotFoundError(err)
		}
		return entities.User{}, err
	}

	return newUserFromModel(user), nil
}

func (r *PostgresRepository) GetUser(ctx context.Context, userID int) (entities.User, error) {
	query := queries.GetUserByID(userID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.User{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.User{}, err
	}
	defer stmt.Close()

	var user models.User
	rows := stmt.QueryRowContext(ctx, args...)
	user, err = scanUser(rows)
	if err != nil {
		return entities.User{}, err
	}

	return newUserFromModel(user), nil
}

func (r *PostgresRepository) GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error) {
	sql, args, err := queries.GetUsers(pageScope).ToSql()
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
			return nil, entities.PageScope{}, errs.NewScanError(err)
		}
		users = append(users, newUserFromModel(usr))
	}
	pageScope.Count = len(users)
	return users, pageScope, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user entities.User) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	userModel := newUserFromEntity(user, false)

	sql, args, err := queries.InsertUser(userModel).ToSql()
	if err != nil {
		level.Error(r.logger).Log("message", "error formating query", "err", err)
		return -1, errs.NewInternalError(err)
	}

	stmt, err := tx.PrepareContext(ctx, sql)
	if err != nil {
		return -1, errs.NewInternalError(err)
	}
	defer stmt.Close()

	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			level.Error(r.logger).Log("message", "error inserting user", "err", err)
			return -1, errs.NewDuplicateEntryError(err)
		}
		level.Error(r.logger).Log("message", "error inserting user", "err", err)
		return -1, errs.NewInternalError(err)
	}
	err = r.AddRoleToUser(ctx, tx, int(lastInsertedID), entities.RoleIDFromName(user.Roles[0]))
	if err != nil {
		return -1, errs.NewInternalError(err)
	}

	if err := tx.Commit(); err != nil {
		level.Error(r.logger).Log("message", "error commiting tx", "err", err)
		return -1, err
	}

	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, userID int, user entities.User) error {
	sql, args, err := queries.UpdateUserInfo(user, userID).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errs.NewNotFoundError(err)
	}

	return nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, userID int) error {
	sql, args, err := queries.DeleteUser(userID).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errs.NewNotFoundError(err)
	}

	return nil
}

func (r *PostgresRepository) GetUserRoles(ctx context.Context, userID int) ([]string, error) {
	sql, args, err := queries.GetRolesByUserID(userID).ToSql()
	if err != nil {
		return nil, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles = make([]string, 0)
	var role string
	for rows.Next() {
		err = rows.Scan(&role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRepository) AddRoleToUser(ctx context.Context, tx *sql.Tx, userID, role int) error {
	sql, args, err := queries.AddRoleToUser(userID, role).ToSql()
	if err != nil {
		return err
	}
	if tx != nil {
		return prepareAndExecute(ctx, tx, sql, args, r.logger)
	}
	return prepareAndExecute(ctx, r.db, sql, args, r.logger)
}

func prepareAndExecute(ctx context.Context, p preparer, sql string, args []interface{}, log log.Logger) error {
	stmt, err := p.PrepareContext(ctx, sql)
	if err != nil {
		level.Error(log).Log("message", "what up preparing", "err", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		level.Error(log).Log("message", "what up executing", "err", err)
		return errs.NewInternalError(err)
	}
	return nil
}

func scanUser(row scannable) (models.User, error) {
	result := models.User{}
	err := row.Scan(
		&result.ID,
		&result.CI,
		&result.Username,
		&result.FirstName,
		&result.LastName,
		&result.DateOfBirth,
		&result.Gender,
		&result.EducationLevel,
		&result.ProviderCode,
		&result.Address,
		&result.CreatedAt,
	)

	return result, err
}
