package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eaguilar88/deu/internal/course_cycle_close_requests"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"go.uber.org/zap"
)

func (r *PostgresRepository) HasPendingCloseRequestForCycle(ctx context.Context, cycleID int64) (bool, error) {
	query, args, err := queries.CountPendingCloseRequestsForCycle(cycleID).ToSql()
	if err != nil {
		return false, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var count int
	if err := stmt.QueryRowContext(ctx, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresRepository) CreateCourseCycleCloseRequest(ctx context.Context, cycleID, submittedByID int64) (int64, error) {
	query, args, err := queries.InsertCourseCycleCloseRequest(cycleID, submittedByID).ToSql()
	if err != nil {
		r.logger.Error("error creating cycle close request query", zap.Error(err))
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("error preparing cycle close request query", zap.Error(err))
		return -1, err
	}
	defer stmt.Close()
	var id int64
	if err = stmt.QueryRowContext(ctx, args...).Scan(&id); err != nil {
		r.logger.Error("error inserting cycle close request", zap.Error(err))
		return -1, err
	}
	return id, nil
}

func (r *PostgresRepository) GetCourseCycleCloseRequestByID(ctx context.Context, id string) (entities.CourseCycleCloseRequest, error) {
	query, args, err := queries.GetCourseCycleCloseRequestByID(id).ToSql()
	if err != nil {
		return entities.CourseCycleCloseRequest{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.CourseCycleCloseRequest{}, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, args...)
	m, err := scanCourseCycleCloseRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.CourseCycleCloseRequest{}, fmt.Errorf("%w: %w", course_cycle_close_requests.ErrCycleCloseRequestNotFound, err)
		}
		return entities.CourseCycleCloseRequest{}, err
	}
	return newCourseCycleCloseRequestFromModel(m), nil
}

func (r *PostgresRepository) GetCourseCycleCloseRequests(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseCycleCloseRequest, entities.PageScope, error) {
	query, args, err := queries.GetCourseCycleCloseRequests(faculty.String(), uint64(pageScope.PerPage), uint64(pageScope.Offset())).ToSql()
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
	var requests []entities.CourseCycleCloseRequest
	for rows.Next() {
		m, err := scanCourseCycleCloseRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		requests = append(requests, newCourseCycleCloseRequestFromModel(m))
	}
	pageScope.Count = len(requests)
	return requests, pageScope, nil
}

func (r *PostgresRepository) ApproveCourseCycleCloseRequest(ctx context.Context, id, reviewerID, cycleID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	approveQuery, approveArgs, err := queries.ApproveCourseCycleCloseRequest(id, reviewerID).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, approveQuery, approveArgs...); err != nil {
		r.logger.Error("error approving cycle close request", zap.Error(err))
		return err
	}

	closeQuery, closeArgs, err := queries.SetCourseCycleClosed(cycleID).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, closeQuery, closeArgs...); err != nil {
		r.logger.Error("error closing course cycle", zap.Error(err))
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) RejectCourseCycleCloseRequest(ctx context.Context, id, reviewerID, comments string) error {
	query, args, err := queries.RejectCourseCycleCloseRequest(id, reviewerID, comments).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w: %w", course_cycle_close_requests.ErrCycleCloseRequestNotFound, err)
	}
	return nil
}

func scanCourseCycleCloseRequest(row scannable) (models.CourseCycleCloseRequest, error) {
	var m models.CourseCycleCloseRequest
	err := row.Scan(
		&m.ID,
		&m.CourseCycleID,
		&m.SubmittedBy,
		&m.Status,
		&m.Comments,
		&m.ReviewerID,
		&m.ReviewedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	return m, err
}

func newCourseCycleCloseRequestFromModel(m models.CourseCycleCloseRequest) entities.CourseCycleCloseRequest {
	req := entities.CourseCycleCloseRequest{
		ID:            m.ID,
		CourseCycleID: m.CourseCycleID,
		Status:        entities.RequestStatus(m.Status),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.Comments.Valid {
		req.Comments = m.Comments.String
	}
	if m.ReviewedAt.Valid {
		req.ReviewedAt = m.ReviewedAt.String
	}
	return req
}
