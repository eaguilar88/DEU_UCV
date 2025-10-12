package repository

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) GetCourseRequest(ctx context.Context, requestID string) (entities.CourseRequest, error) {
	query := queries.GetCourseRequestByID(requestID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.CourseRequest{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.CourseRequest{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return entities.CourseRequest{}, err
	}
	defer rows.Close()
	var request models.CourseRequest
	for rows.Next() {
		request, err = scanCourseRequest(rows)
		if err != nil {
			return entities.CourseRequest{}, err
		}
	}
	return newCourseRequestFromModel(request), nil
}

func (r *PostgresRepository) GetCourseRequests(ctx context.Context, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error) {
	sql, args, err := queries.GetCourseRequests(pageScope).ToSql()
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

	var requests []entities.CourseRequest
	for rows.Next() {
		request, err := scanCourseRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		requests = append(requests, newCourseRequestFromModel(request))
	}
	pageScope.Count = len(requests)
	return requests, pageScope, nil
}

func (r *PostgresRepository) CreateCourseRequest(ctx context.Context, request entities.CourseRequest) (int64, error) {
	panic("")
}

func (r *PostgresRepository) UpdateCourseRequest(ctx context.Context, requestID string, request entities.CourseRequest) error {
	panic("")
}

func (r *PostgresRepository) DeleteCourseRequest(ctx context.Context, requestID string) error {
	panic("")
}

func scanCourseRequest(row scannable) (models.CourseRequest, error) {
	result := models.CourseRequest{}
	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.Type,
		&result.Status,
		&result.Comments,
		&result.UserID,
		&result.UserFirstName,
		&result.UserLastName,
		&result.ReviewerID,
		&result.ReviewerFirstName,
		&result.ReviewerLastName,
		&result.ReviewedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	return result, err
}

func newCourseRequestFromModel(request models.CourseRequest) entities.CourseRequest {
	result := entities.CourseRequest{
		ID: request.ID,
		User: entities.User{
			ID:        request.UserID,
			FirstName: request.UserFirstName,
			LastName:  request.UserLastName,
		},
		Reviewer: entities.User{
			ID:        request.ReviewerID,
			FirstName: request.ReviewerFirstName,
			LastName:  request.ReviewerLastName,
		},
		Status:      entities.RequestStatus(request.Status),
		Type:        request.Type,
		ReviewedAt:  request.ReviewedAt,
		CreatedAt:   request.CreatedAt,
		UpdatedAtAt: request.UpdatedAt,
	}

	if request.Name.Valid {
		result.Name = request.Name.String
	}

	if request.Description.Valid {
		result.Description = request.Description.String
	}

	if request.Comments.Valid {
		result.Comments = request.Comments.String
	}

	return result
}
