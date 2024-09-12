package repository

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
	"github.com/eaguilar88/deu/pkg/repository/queries"
)

func (r *PostgresRepository) GetEndorsement(ctx context.Context, endorsementID int) (entities.Endorsements, error) {
	query := queries.GetEndorsementByID(endorsementID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.Endorsements{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.Endorsements{}, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return entities.Endorsements{}, err
	}
	defer rows.Close()
	var endorsement models.Endorsement
	for rows.Next() {
		endorsement, err = scanEndorsment(rows)
		if err != nil {
			return entities.Endorsements{}, err
		}
	}
	return newEndorsmentFromModel(endorsement), nil
}
func (r *PostgresRepository) GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.Endorsements, entities.PageScope, error) {
	panic("")
}
func (r *PostgresRepository) CreateEndorsement(ctx context.Context, endorsement entities.Endorsements) (int64, error) {
	panic("")
}
func (r *PostgresRepository) UpdateEndorsement(ctx context.Context, endorsementID int, endorsement entities.Endorsements) error {
	panic("")
}
func (r *PostgresRepository) DeleteEndorsement(ctx context.Context, endorsementID int) error {
	panic("")
}

func scanEndorsment(row scannable) (models.Endorsement, error) {
	result := models.Endorsement{}
	err := row.Scan(
		&result.ID,
		&result.UserID,
		&result.Status,
		&result.Path,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	return result, err
}
