package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	errs "github.com/eaguilar88/deu/internal/errors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetGroupRequestByID(ctx context.Context, requestID string) (entities.GroupAuthRequest, error) {
	sql, args, err := queries.GetGroupRequestByID(requestID).ToSql()
	if err != nil {
		return entities.GroupAuthRequest{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.GroupAuthRequest{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return entities.GroupAuthRequest{}, err
	}
	defer rows.Close()
	var request models.GroupAuthRequest
	for rows.Next() {
		request, err = scanGroupRequest(rows)
		if err != nil {
			return entities.GroupAuthRequest{}, err
		}
	}
	return newGroupRequestFromModel(request), nil
}

func (r *PostgresRepository) GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, scope entities.PageScope) ([]entities.GroupAuthRequest, entities.PageScope, error) {
	sql, args, err := queries.GetGroupRequestsByFaculty(faculty.String(), scope.PerPage, scope.Offset()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	r.logger.Debug("SQL Query: ", zap.String("query", sql), zap.Any("args", args))
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
	var requests []entities.GroupAuthRequest
	for rows.Next() {
		model, err := scanGroupRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		requests = append(requests, newGroupRequestFromModel(model))
	}
	var total int
	sql, args, err = queries.CountGroupRequestsByFaculty(faculty.String()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	if err = r.db.QueryRowContext(ctx, sql, args...).Scan(&total); err != nil {
		return nil, entities.PageScope{}, err
	}
	scope.Count = total
	return requests, scope, nil
}

func (r *PostgresRepository) ApproveGroupRequest(ctx context.Context, reqID string) error {
	sql, args, err := queries.ApproveGroupRequest(reqID).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
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
		return errs.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) RejectGroupRequest(ctx context.Context, reqID string) error {
	sql, args, err := queries.RejectGroupRequest(reqID).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
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
		return errs.ErrNotFound
	}
	return nil
}

func newGroupRequestFromModel(m models.GroupAuthRequest) entities.GroupAuthRequest {
	gar := entities.GroupAuthRequest{
		ID:        fmt.Sprintf("%d", m.ID),
		GroupID:   fmt.Sprintf("%d", m.GroupID),
		Faculty:   entities.Faculty(m.Faculty),
		Status:    entities.RequestStatus(m.Status),
		Comments:  m.Comments,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}

	if m.ReviewerID.Valid {
		gar.Reviewer = &entities.User{
			ID: fmt.Sprintf("%d", m.ReviewerID.Int64),
		}
	}

	return gar
}

func scanGroupRequest(row scannable) (models.GroupAuthRequest, error) {
	result := models.GroupAuthRequest{}
	err := row.Scan(
		&result.ID,
		&result.GroupID,
		&result.Faculty,
		&result.Status,
		&result.Comments,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.ReviewerID,
		&result.ReviewedAt,
	)
	if err != nil {
		return models.GroupAuthRequest{}, err
	}
	return result, nil
}
