package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	errs "github.com/eaguilar88/deu/internal/errors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
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
		&user.Email,
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

func (r *PostgresRepository) GetUser(ctx context.Context, userID string) (entities.User, error) {
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
	sql, args, err := queries.GetUsers(pageScope.PerPage, pageScope.Offset()).ToSql()
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
		r.logger.Error("error formatting query", zap.Error(err))
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
			r.logger.Error("error inserting user", zap.Error(err))
			return -1, errs.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting user", zap.Error(err))
		return -1, errs.NewInternalError(err)
	}
	err = r.AddRoleToUser(
		ctx,
		tx,
		fmt.Sprintf("%d", lastInsertedID),
		entities.RoleIDFromName(user.Roles[0]),
	)
	if err != nil {
		return -1, errs.NewInternalError(err)
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("error committing tx", zap.Error(err))
		return -1, err
	}

	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, userID string, user entities.User) error {
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

func (r *PostgresRepository) DeleteUser(ctx context.Context, userID string) error {
	sql, args, err := queries.SoftDeleteUser(userID).ToSql()
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

func (r *PostgresRepository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
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

	roles := make([]string, 0)
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

func (r *PostgresRepository) AddRoleToUser(ctx context.Context, tx *sql.Tx, userID string, role int) error {
	sql, args, err := queries.AddRoleToUser(userID, role).ToSql()
	if err != nil {
		return err
	}
	if tx != nil {
		return prepareAndExecute(ctx, tx, sql, args, r.logger)
	}
	return prepareAndExecute(ctx, r.db, sql, args, r.logger)
}

func (r *PostgresRepository) GetUsersByCoursePeriodID(
	ctx context.Context,
	periodID string,
) ([]entities.User, error) {
	sql, args, err := queries.GetUsersByCoursePeriodID(periodID).ToSql()
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

	var users []entities.User
	for rows.Next() {
		result := models.User{}
		err = rows.Scan(
			&result.ID,
			&result.CI,
			&result.FirstName,
			&result.LastName,
		)
		if err != nil {
			return nil, errs.NewScanError(err)
		}
		users = append(users, newUserFromModel(result))
	}
	return users, nil
}

func prepareAndExecute(
	ctx context.Context,
	p preparer,
	sql string,
	args []interface{},
	log *zap.Logger,
) error {
	stmt, err := p.PrepareContext(ctx, sql)
	if err != nil {
		log.Error("error preparing", zap.Error(err))
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		log.Error("error executing", zap.Error(err))
		return errs.NewInternalError(err)
	}
	return nil
}

func scanUser(row scannable) (models.User, error) {
	result := models.User{}
	err := row.Scan(
		&result.ID,
		&result.CI,
		&result.Email,
		&result.FirstName,
		&result.LastName,
		&result.DateOfBirth,
		&result.Gender,
		&result.EducationLevel,
		&result.Address,
		&result.CreatedAt,
		&result.ProviderCode,
	)

	return result, err
}

func newUserFromEntity(user entities.User, isUpdate bool) models.User {
	ci, err := strconv.Atoi(user.CI)
	if err != nil {
		ci = 0
	}
	model := models.User{
		ID:        user.ID,
		CI:        ci,
		Email:     user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		DateOfBirth: sql.NullString{
			String: user.DateOfBirth,
			Valid:  true,
		},
		Gender: sql.NullString{
			String: user.Gender,
			Valid:  true,
		},
		EducationLevel: user.EducationLevel,
		Address: sql.NullString{
			String: user.Address,
			Valid:  true,
		},
		Password:  user.Password,
		CreatedAt: time.Now().String(),
		UpdatedAt: time.Now().String(),
	}
	if isUpdate {
		model.CreatedAt = user.CreatedAt
	}
	return model
}

func newUserFromModel(user models.User) entities.User {
	entity := entities.User{
		ID:             user.ID,
		CI:             fmt.Sprintf("%d", user.CI),
		Username:       user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		EducationLevel: user.EducationLevel,
		Password:       user.Password,
		CreatedAt:      user.CreatedAt,
	}

	if user.Gender.Valid {
		entity.Gender = user.Gender.String
	}
	if user.Address.Valid {
		entity.Address = user.Address.String
	}

	if user.DateOfBirth.Valid {
		entity.DateOfBirth = user.DateOfBirth.String
	}

	if user.ProviderCode.Valid {
		entity.ProviderCode = user.ProviderCode.String
	}

	entity.SetAge()
	return entity
}
