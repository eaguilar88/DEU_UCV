package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error) {
	query, args, err := queries.GetGroupByID(groupID).ToSql()
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	defer stmt.Close()
	var group models.ExtensionGroup
	row := stmt.QueryRowContext(ctx, args...)
	group, err = scanGroup(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.ExtensionGroup{}, fmt.Errorf("%w: %w", groups.ErrGroupNotFound, err)
		}
		return entities.ExtensionGroup{}, err
	}
	return newGroupFromModel(group), nil
}

func (r *PostgresRepository) GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
	query, args, err := queries.GetGroups(pageScope.PerPage, pageScope.Offset()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
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
	query, args, err := queries.InsertGroup(newGroupToModel(gr)).ToSql()
	if err != nil {
		return 0, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("error inserting group", zap.Error(err))
			return -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting group", zap.Error(err))
		return -1, httperrors.NewInternalError(err)
	}
	return lastInsertedID, nil
}

// CreateGroupWithRequest creates a group and its authorization request in a single transaction
func (r *PostgresRepository) CreateGroupWithRequest(ctx context.Context, group entities.ExtensionGroup, request entities.GroupRequest) (groupID int64, requestID int64, err error) {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if commit is not called

	// 1. Create the group
	groupSQL, groupArgs, err := queries.InsertGroup(newGroupToModel(group)).ToSql()
	if err != nil {
		r.logger.Error("failed to build group query", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to build group query: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, groupSQL)
	if err != nil {
		r.logger.Error("failed to prepare group statement", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to prepare group statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, groupArgs...).Scan(&groupID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("duplicate group entry", zap.Error(err))
			return -1, -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("failed to insert group", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to insert group: %w", err)
	}

	// 2. Create the group authorization request
	requestSQL, requestArgs, err := queries.InsertGroupRequest(models.GroupRequest{
		GroupID: groupID,
		Status:  string(request.Status),
		Faculty: string(request.Faculty),
		Comments: sql.NullString{
			String: request.Comments,
			Valid:  request.Comments != "",
		},
	}).ToSql()
	if err != nil {
		r.logger.Error("failed to build request query", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to build request query: %w", err)
	}

	stmt2, err := tx.PrepareContext(ctx, requestSQL)
	if err != nil {
		r.logger.Error("failed to prepare request statement", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to prepare request statement: %w", err)
	}
	defer stmt2.Close()

	err = stmt2.QueryRowContext(ctx, requestArgs...).Scan(&requestID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("duplicate group request entry", zap.Error(err))
			return -1, -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("failed to insert group request", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to insert group request: %w", err)
	}

	// 3. Commit the transaction
	if err := tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("group and request created successfully",
		zap.Int64("group_id", groupID),
		zap.Int64("request_id", requestID))

	return groupID, requestID, nil
}

func (r *PostgresRepository) UpdateGroup(ctx context.Context, group entities.ExtensionGroup) error {
	query, args, err := queries.UpdateGroup(newGroupToModel(group)).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w: %w", groups.ErrGroupNotFound, err)
	}
	return nil
}

func (r *PostgresRepository) DeleteGroup(ctx context.Context, groupID string) error {
	query, args, err := queries.DeleteGroup(groupID).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w: %w", groups.ErrGroupNotFound, err)
	}
	return nil
}

func (r *PostgresRepository) CreateGroupRequest(ctx context.Context, req entities.GroupRequest) (int64, error) {
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

	query, args, err := queries.InsertGroupRequest(models.GroupRequest{
		GroupID: int64(groupID),
		Status:  string(req.Status),
		Faculty: string(req.Faculty),
		Comments: sql.NullString{
			String: req.Comments,
			Valid:  req.Comments != "",
		},
	}).ToSql()
	if err != nil {
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("error inserting group request", zap.Error(err), zap.String("group_id", req.GroupID))
			return -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting group request", zap.Error(err), zap.String("group_id", req.GroupID))
		return -1, httperrors.NewInternalError(err)
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
		&group.UserID,
		&group.Name,
		&group.Description,
		&group.Faculty,
		&group.Objective,
		&group.Code,
		&group.Director,
		&group.Type,
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
		DeletedAt: group.DeletedAt.String,
	}
}
