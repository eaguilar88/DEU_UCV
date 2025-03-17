package repository

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
)

func (r *PostgresRepository) GetCoursePeriodByID(ctx context.Context, periodID int) (entities.CoursePeriod, error) {
	return entities.CoursePeriod{}, nil
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
