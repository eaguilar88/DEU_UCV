package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/eaguilar88/deu/internal/provider_requests"
	"go.uber.org/zap"
)

func (r *PostgresRepository) CreateProviderRequest(ctx context.Context, providerID int64) error {
	query, args, err := queries.InsertProviderRequest(providerID).ToSql()
	if err != nil {
		r.logger.Error("error creating provider request query", zap.Error(err))
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		r.logger.Error("error preparing provider request query", zap.Error(err))
		return err
	}
	defer stmt.Close()
	var id int64
	if err = stmt.QueryRowContext(ctx, args...).Scan(&id); err != nil {
		r.logger.Error("error inserting provider request", zap.Error(err))
		return err
	}
	return nil
}

func (r *PostgresRepository) GetProviderRequestByID(ctx context.Context, id string) (entities.ProviderRequest, error) {
	query, args, err := queries.GetProviderRequestByID(id).ToSql()
	if err != nil {
		return entities.ProviderRequest{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.ProviderRequest{}, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, args...)
	m, err := scanProviderRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.ProviderRequest{}, fmt.Errorf("%w: %w", provider_requests.ErrProviderRequestNotFound, err)
		}
		return entities.ProviderRequest{}, err
	}
	return newProviderRequestFromModel(m), nil
}

func (r *PostgresRepository) GetProviderRequests(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.ProviderRequest, entities.PageScope, error) {
	query, args, err := queries.GetProviderRequests(faculty.String(), uint64(pageScope.PerPage), uint64(pageScope.Offset())).ToSql()
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
	var requests []entities.ProviderRequest
	for rows.Next() {
		m, err := scanProviderRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		requests = append(requests, newProviderRequestFromModel(m))
	}
	pageScope.Count = len(requests)
	return requests, pageScope, nil
}

func (r *PostgresRepository) ApproveProviderRequest(ctx context.Context, id, reviewerID string, providerID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	approveQuery, approveArgs, err := queries.ApproveProviderRequest(id, reviewerID).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, approveQuery, approveArgs...); err != nil {
		r.logger.Error("error approving provider request", zap.Error(err))
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) RejectProviderRequest(ctx context.Context, id, reviewerID, comments string) error {
	query, args, err := queries.RejectProviderRequest(id, reviewerID, comments).ToSql()
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
		return fmt.Errorf("%w: %w", provider_requests.ErrProviderRequestNotFound, err)
	}
	return nil
}

func scanProviderRequest(row scannable) (models.ProviderRequest, error) {
	var m models.ProviderRequest
	err := row.Scan(
		&m.ID,
		&m.ProviderID,
		&m.Status,
		&m.ReviewerID,
		&m.Comments,
		&m.ReviewedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	return m, err
}

func newProviderRequestFromModel(m models.ProviderRequest) entities.ProviderRequest {
	req := entities.ProviderRequest{
		ID:         m.ID,
		ProviderID: m.ProviderID,
		Status:     entities.RequestStatus(m.Status),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
	if m.Comments.Valid {
		req.Comments = m.Comments.String
	}
	if m.ReviewedAt.Valid {
		req.ReviewedAt = m.ReviewedAt.String
	}
	return req
}
