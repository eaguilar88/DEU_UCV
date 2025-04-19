package repository

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) GetEndorsement(ctx context.Context, endorsementID string) (entities.Endorsement, error) {
	query := queries.GetEndorsementByID(endorsementID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.Endorsement{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.Endorsement{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return entities.Endorsement{}, err
	}
	defer rows.Close()
	var endorsement models.EndorsementRequest
	for rows.Next() {
		endorsement, err = scanEndorsment(rows)
		if err != nil {
			return entities.Endorsement{}, err
		}
	}
	return newEndorsmentFromModel(endorsement), nil
}
func (r *PostgresRepository) GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.Endorsement, entities.PageScope, error) {
	sql, args, err := queries.GetEndorsements(pageScope).ToSql()
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

	var endorsements []entities.Endorsement
	for rows.Next() {
		endorsement, err := scanEndorsment(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		endorsements = append(endorsements, newEndorsmentFromModel(endorsement))
	}
	pageScope.Count = len(endorsements)
	return endorsements, pageScope, nil
}

func (r *PostgresRepository) CreateEndorsement(ctx context.Context, endorsement entities.Endorsement) (int64, error) {
	panic("")
}
func (r *PostgresRepository) UpdateEndorsement(ctx context.Context, endorsementID string, endorsement entities.Endorsement) error {
	panic("")
}
func (r *PostgresRepository) DeleteEndorsement(ctx context.Context, endorsementID string) error {
	panic("")
}

func scanEndorsment(row scannable) (models.EndorsementRequest, error) {
	result := models.EndorsementRequest{}
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

func newEndorsmentFromModel(endorsement models.EndorsementRequest) entities.Endorsement {
	result := entities.Endorsement{
		ID: endorsement.ID,
		User: entities.User{
			ID:        endorsement.UserID,
			FirstName: endorsement.UserFirstName,
			LastName:  endorsement.UserLastName,
		},
		Reviewer: entities.User{
			ID:        endorsement.ReviewerID,
			FirstName: endorsement.ReviewerFirstName,
			LastName:  endorsement.ReviewerLastName,
		},
		Status:      entities.EndorsementStatus(endorsement.Status),
		Type:        endorsement.Type,
		ReviewedAt:  endorsement.ReviewedAt,
		CreatedAt:   endorsement.CreatedAt,
		UpdatedAtAt: endorsement.UpdatedAt,
	}

	if endorsement.Name.Valid {
		result.Name = endorsement.Name.String
	}

	if endorsement.Description.Valid {
		result.Description = endorsement.Description.String
	}

	if endorsement.Comments.Valid {
		result.Comments = endorsement.Comments.String
	}

	return result
}
