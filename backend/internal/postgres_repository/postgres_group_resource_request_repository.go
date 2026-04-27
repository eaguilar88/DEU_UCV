package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/group_resource_requests"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) CreateGroupResourceRequest(ctx context.Context, req entities.GroupResourceRequest) (int64, error) {
	model := models.GroupResourceRequest{
		Type:   req.Type,
		Status: req.Status,
	}
	if req.GroupID != "" {
		if _, err := fmt.Sscanf(req.GroupID, "%d", &model.GroupID); err != nil {
			return 0, err
		}
	}
	if req.Content != "" {
		model.Content.String = req.Content
		model.Content.Valid = true
	}

	query, args, err := queries.InsertGroupResourceRequest(model).ToSql()
	if err != nil {
		return 0, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var id int64
	if err := stmt.QueryRowContext(ctx, args...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *PostgresRepository) GetGroupResourceRequestByID(ctx context.Context, reqID string) (entities.GroupResourceRequest, error) {
	query, args, err := queries.GetGroupResourceRequestByID(reqID).ToSql()
	if err != nil {
		return entities.GroupResourceRequest{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.GroupResourceRequest{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return entities.GroupResourceRequest{}, err
	}
	defer rows.Close()
	found := false
	var m models.GroupResourceRequest
	for rows.Next() {
		found = true
		m, err = scanGroupResourceRequest(rows)
		if err != nil {
			return entities.GroupResourceRequest{}, err
		}
	}
	if !found {
		return entities.GroupResourceRequest{}, fmt.Errorf("%w", group_resource_requests.ErrNotFound)
	}
	return newGroupResourceRequestFromModel(m), nil
}

func (r *PostgresRepository) GetGroupResourceRequestsByFaculty(ctx context.Context, faculty entities.Faculty, scope entities.PageScope) ([]entities.GroupResourceRequest, entities.PageScope, error) {
	query, args, err := queries.GetGroupResourceRequestsByFaculty(faculty.String(), scope.PerPage, scope.Offset()).ToSql()
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
	var reqs []entities.GroupResourceRequest
	for rows.Next() {
		m, err := scanGroupResourceRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		reqs = append(reqs, newGroupResourceRequestFromModel(m))
	}
	var total int
	cQuery, cArgs, err := queries.CountGroupResourceRequestsByFaculty(faculty.String()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	if err := r.db.QueryRowContext(ctx, cQuery, cArgs...).Scan(&total); err != nil {
		return nil, entities.PageScope{}, err
	}
	scope.Count = total
	return reqs, scope, nil
}

func (r *PostgresRepository) ApproveGroupResourceRequest(ctx context.Context, reqID string) error {
	query, args, err := queries.ApproveGroupResourceRequest(reqID).ToSql()
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
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w", group_resource_requests.ErrNotFound)
	}
	return nil
}

func (r *PostgresRepository) RejectGroupResourceRequest(ctx context.Context, reqID string) error {
	query, args, err := queries.RejectGroupResourceRequest(reqID).ToSql()
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
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w", group_resource_requests.ErrNotFound)
	}
	return nil
}

func scanGroupResourceRequest(row scannable) (models.GroupResourceRequest, error) {
	var m models.GroupResourceRequest
	err := row.Scan(&m.ID, &m.GroupID, &m.Type, &m.Content, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func newGroupResourceRequestFromModel(m models.GroupResourceRequest) entities.GroupResourceRequest {
	return entities.GroupResourceRequest{
		ID:        fmt.Sprintf("%d", m.ID),
		GroupID:   fmt.Sprintf("%d", m.GroupID),
		Type:      m.Type,
		Content:   m.Content.String,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}
