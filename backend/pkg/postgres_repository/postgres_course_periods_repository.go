package repository

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
	"github.com/eaguilar88/deu/pkg/postgres_repository/queries"
)

func (r *PostgresRepository) GetCoursePeriodByID(ctx context.Context, periodID int) (entities.CoursePeriod, error) {
	query := queries.GetCoursePeriodByID(periodID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	defer stmt.Close()
	var coursePeriod models.CoursePeriod
	row := stmt.QueryRowContext(ctx, args...)
	coursePeriod, err = scanCoursePeriod(row)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	return newCoursePeriodFromModel(coursePeriod), nil
}

func (r *PostgresRepository) GetCoursePeriods(ctx context.Context, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
	return []entities.CoursePeriod{}, entities.PageScope{}, nil
}

func (r *PostgresRepository) CreateCoursePeriod(ctx context.Context, coursePeriod entities.CoursePeriod) (int64, error) {
	return 0, nil
}

func (r *PostgresRepository) UpdateCoursePeriod(ctx context.Context, periodID int, coursePeriod entities.CoursePeriod) error {
	return nil
}

func (r *PostgresRepository) DeleteCoursePeriod(ctx context.Context, periodID int) error {
	return nil
}

func scanCoursePeriod(row scannable) (models.CoursePeriod, error) {
	var coursePeriod models.CoursePeriod
	err := row.Scan(
		&coursePeriod.ID,
		&coursePeriod.Course.ID,
		&coursePeriod.Course.Name,
		&coursePeriod.Course.Description,
		&coursePeriod.StartDate,
		&coursePeriod.EndDate,
		&coursePeriod.IsActive,
		&coursePeriod.InscriptionDate,
		&coursePeriod.CreatedAt,
		&coursePeriod.UpdatedAt,
		&coursePeriod.DeletedAt,
	)
	return coursePeriod, err
}

func newCoursePeriodFromModel(coursePeriod models.CoursePeriod) entities.CoursePeriod {
	var description, deletedAt string
	if coursePeriod.Course.Description.Valid {
		description = coursePeriod.Course.Description.String
	}

	if coursePeriod.DeletedAt.Valid {
		deletedAt = coursePeriod.DeletedAt.String
	}

	return entities.CoursePeriod{
		ID: coursePeriod.ID,
		Course: entities.Course{
			ID:          coursePeriod.Course.ID,
			Name:        coursePeriod.Course.Name,
			Description: description,
		},
		StartDate:       coursePeriod.StartDate,
		EndDate:         coursePeriod.EndDate,
		InscriptionDate: coursePeriod.InscriptionDate,
		IsActive:        coursePeriod.IsActive,
		CreatedAt:       coursePeriod.CreatedAt,
		UpdatedAt:       coursePeriod.UpdatedAt,
		DeletedAt:       deletedAt,
	}
}
