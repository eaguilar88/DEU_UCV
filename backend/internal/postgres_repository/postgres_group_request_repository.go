package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/eaguilar88/deu/internal/users"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetGroupRequestByID(ctx context.Context, requestID string) (entities.GroupRequest, error) {
	query, args, err := queries.GetGroupRequestByID(requestID).ToSql()
	if err != nil {
		return entities.GroupRequest{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.GroupRequest{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return entities.GroupRequest{}, err
	}
	defer rows.Close()
	var request models.GroupRequest
	found := false
	for rows.Next() {
		found = true
		request, err = scanGroupRequest(rows)
		if err != nil {
			return entities.GroupRequest{}, err
		}
	}
	if !found {
		return entities.GroupRequest{}, fmt.Errorf("%w", group_requests.ErrGroupRequestNotFound)
	}
	return newGroupRequestFromModel(request), nil
}

func (r *PostgresRepository) GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, scope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error) {
	query, args, err := queries.GetGroupRequestsByFaculty(faculty.String(), status, scope.PerPage, scope.Offset()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, 0, err
	}
	r.logger.Debug("SQL Query: ", zap.String("query", query), zap.Any("args", args))
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, entities.PageScope{}, 0, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, entities.PageScope{}, 0, err
	}
	defer rows.Close()
	var requests []entities.GroupRequest
	for rows.Next() {
		model, err := scanGroupRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, 0, err
		}
		requests = append(requests, newGroupRequestFromModel(model))
	}
	var total int
	query, args, err = queries.CountGroupRequestsByFaculty(faculty.String(), status).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, 0, err
	}
	if err = r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return nil, entities.PageScope{}, 0, err
	}
	scope.Count = total

	var pendingCount int
	query, args, err = queries.CountPendingGroupRequestsByFaculty(faculty.String()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, 0, err
	}
	if err = r.db.QueryRowContext(ctx, query, args...).Scan(&pendingCount); err != nil {
		return nil, entities.PageScope{}, 0, err
	}

	return requests, scope, pendingCount, nil
}

func (r *PostgresRepository) GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error) {
	query, args, err := queries.GetPendingGroupRequestsCountGroupedByFaculty(faculty.String()).ToSql()
	if err != nil {
		return nil, err
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entities.FacultyPendingCount
	for rows.Next() {
		var f string
		var count int
		if err := rows.Scan(&f, &count); err != nil {
			return nil, err
		}
		fac, err := entities.FromString(f)
		if err != nil {
			return nil, err
		}
		result = append(result, entities.FacultyPendingCount{
			Faculty: fac,
			Count:   count,
		})
	}
	return result, nil
}

func (r *PostgresRepository) GetGroupRequestsByGroupID(ctx context.Context, groupID string) ([]entities.GroupRequest, error) {
	query, args, err := queries.GetGroupRequestsByGroupID(groupID).ToSql()
	if err != nil {
		return nil, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var requests []entities.GroupRequest
	for rows.Next() {
		model, err := scanGroupRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, newGroupRequestFromModel(model))
	}
	return requests, nil
}

func (r *PostgresRepository) ApproveGroupRequest(ctx context.Context, reqID string) error {
	query, args, err := queries.ApproveGroupRequest(reqID).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", group_requests.ErrGroupRequestNotFound)
	}
	return nil
}

func (r *PostgresRepository) RejectGroupRequest(ctx context.Context, reqID string) error {
	query, args, err := queries.RejectGroupRequest(reqID).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", group_requests.ErrGroupRequestNotFound)
	}
	return nil
}

// CreateGroupAdminAndActivate creates the login user for a group's admin, assigns it a role,
// links it to the group, and activates the group — all inside a single transaction, so a
// failure partway through never leaves an orphaned, unlinked user behind.
func (r *PostgresRepository) CreateGroupAdminAndActivate(ctx context.Context, groupID string, adminUser entities.User) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return -1, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userModel := newUserFromEntity(adminUser, false)

	userSQL, userArgs, err := queries.InsertUser(userModel).ToSql()
	if err != nil {
		r.logger.Error("error formatting insert user query", zap.Error(err))
		return -1, fmt.Errorf("database error: %w", err)
	}

	userStmt, err := tx.PrepareContext(ctx, userSQL)
	if err != nil {
		return -1, fmt.Errorf("database error: %w", err)
	}
	defer userStmt.Close()

	var newUserID int64
	err = userStmt.QueryRowContext(ctx, userArgs...).Scan(&newUserID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("duplicate group admin user entry", zap.Error(err))
			return -1, fmt.Errorf("%w: %w", users.ErrUserAlreadyExists, err)
		}
		r.logger.Error("failed to insert group admin user", zap.Error(err))
		return -1, fmt.Errorf("database error: %w", err)
	}

	userIDStr := fmt.Sprintf("%d", newUserID)

	if err := r.AddRoleToUser(ctx, tx, userIDStr, entities.RoleIDFromName(adminUser.Roles[0])); err != nil {
		r.logger.Error("failed to assign role to group admin user", zap.Error(err))
		return -1, fmt.Errorf("database error: %w", err)
	}

	updateGroupSQL, updateGroupArgs, err := queries.UpdateGroupUserID(groupID, userIDStr).ToSql()
	if err != nil {
		r.logger.Error("failed to build update group user_id query", zap.Error(err))
		return -1, fmt.Errorf("database error: %w", err)
	}
	if err := prepareAndExecute(ctx, tx, updateGroupSQL, updateGroupArgs, r.logger); err != nil {
		return -1, fmt.Errorf("failed to update group user_id: %w", err)
	}

	activateSQL, activateArgs, err := queries.ActivateGroup(groupID).ToSql()
	if err != nil {
		r.logger.Error("failed to build activate group query", zap.Error(err))
		return -1, fmt.Errorf("database error: %w", err)
	}
	if err := prepareAndExecute(ctx, tx, activateSQL, activateArgs, r.logger); err != nil {
		return -1, fmt.Errorf("failed to activate group: %w", err)
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return -1, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newUserID, nil
}

func newGroupRequestFromModel(m models.GroupRequest) entities.GroupRequest {
	gar := entities.GroupRequest{
		ID:        fmt.Sprintf("%d", m.ID),
		GroupID:   fmt.Sprintf("%d", m.GroupID),
		GroupName: m.GroupName,
		Faculty:   entities.Faculty(m.Faculty),
		Status:    entities.RequestStatus(m.Status),
		Comments:  m.GetComments(),
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}

	if m.ReviewerID.Valid {
		gar.Reviewer = &entities.User{
			ID: fmt.Sprintf("%d", m.ReviewerID.Int64),
		}
	}

	if m.ReviewedAt.Valid {
		gar.ReviewedAt = m.ReviewedAt.String
	}

	return gar
}

func scanGroupRequest(row scannable) (models.GroupRequest, error) {
	result := models.GroupRequest{}
	err := row.Scan(
		&result.ID,
		&result.GroupID,
		&result.GroupName,
		&result.Faculty,
		&result.Status,
		&result.Comments,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.ReviewerID,
		&result.ReviewedAt,
	)
	if err != nil {
		return models.GroupRequest{}, err
	}
	return result, nil
}
