package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
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

func (r *PostgresRepository) ActivateGroup(ctx context.Context, groupID string) error {
	query, args, err := queries.ActivateGroup(groupID).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
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
