package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	errs "github.com/eaguilar88/deu/internal/errors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error) {
	sql, args, err := queries.GetGroupByID(groupID).ToSql()
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	defer stmt.Close()
	var group models.ExtensionGroup
	row := stmt.QueryRowContext(ctx, args...)
	group, err = scanGroup(row)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	return newGroupFromModel(group), nil
}

func (r *PostgresRepository) GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
	sql, args, err := queries.GetGroups(pageScope.PerPage, pageScope.Offset()).ToSql()
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
	var groups []entities.ExtensionGroup
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		groups = append(groups, newGroupFromModel(group))
	}
	return groups, pageScope, nil
}

func (r *PostgresRepository) CreateGroup(ctx context.Context, gr entities.ExtensionGroup) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	sql, args, err := queries.InsertGroup(newGroupToModel(gr)).ToSql()
	if err != nil {
		return 0, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("error inserting group", zap.Error(err))
			return -1, errs.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting group", zap.Error(err))
		return -1, errs.NewInternalError(err)
	}
	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateGroup(ctx context.Context, group entities.ExtensionGroup) error {
	sql, args, err := queries.UpdateGroup(newGroupToModel(group)).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return err
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

func (r *PostgresRepository) DeleteGroup(ctx context.Context, groupID string) error {
	sql, args, err := queries.DeleteGroup(groupID).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return err
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

func (r *PostgresRepository) CreateGroupRequest(ctx context.Context, req entities.GroupAuthRequest) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback()

	groupID, err := strconv.Atoi(req.GroupID)
	if err != nil {
		r.logger.Warn("Non numeric group_id", zap.String("group_id", req.GroupID), zap.Error(err))
		return -1, err
	}

	sql, args, err := queries.InsertGroupRequest(models.GroupAuthRequest{
		GroupID:  int64(groupID),
		Status:   string(req.Status),
		Faculty:  string(req.Faculty),
		Comments: req.Comments,
	}).ToSql()
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
			r.logger.Error("error inserting group request", zap.Error(err), zap.String("group_id", req.GroupID))
			return -1, errs.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting group request", zap.Error(err), zap.String("group_id", req.GroupID))
		return -1, errs.NewInternalError(err)
	}

	if err := tx.Commit(); err != nil {
		return -1, err
	}
	return lastInsertedID, nil
}

func scanGroup(row scannable) (models.ExtensionGroup, error) {
	var group models.ExtensionGroup
	err := row.Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.Faculty,
		&group.Objective,
		&group.Location,
		&group.IsActive,
		&group.CreatedAt,
		&group.UpdatedAt,
		&group.DeletedAt,
	)
	if err != nil {
		return models.ExtensionGroup{}, err
	}
	return group, nil
}

func newGroupToModel(group entities.ExtensionGroup) models.ExtensionGroup {
	return models.ExtensionGroup{
		ID:   group.ID,
		Name: group.Name,
		Description: sql.NullString{
			String: group.Description,
			Valid:  true,
		},
		Faculty:   group.Faculty.String(),
		UserID:    group.Owner.ID,
		Objective: group.Objective,
		Location: sql.NullString{
			String: group.Location,
			Valid:  true,
		},
		IsActive:  group.Active,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
}

func newGroupFromModel(group models.ExtensionGroup) entities.ExtensionGroup {
	var description, objective, location string
	if group.Description.Valid {
		description = group.Description.String
	}

	if group.Location.Valid {
		location = group.Location.String
	}

	return entities.ExtensionGroup{
		ID:          group.ID,
		Name:        group.Name,
		Description: description,
		Owner: &entities.User{
			ID: group.UserID,
		},
		Objective: objective,
		Location:  location,
		Active:    group.IsActive,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
		DeletedAt: group.DeletedAt,
	}
}
