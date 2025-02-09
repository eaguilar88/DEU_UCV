package repository

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
)

func (r *PostgresRepository) GetCourse(ctx context.Context, courseID int) (entities.Course, error) {
	panic("not implemented")
}

func (r *PostgresRepository) GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	panic("not implemented")
}

func (r *PostgresRepository) CreateCourse(ctx context.Context, course entities.Course) (int64, error) {
	panic("not implemented")
}

func (r *PostgresRepository) UpdateCourse(ctx context.Context, courseID int, course entities.Course) error {
	panic("not implemented")
}

func (r *PostgresRepository) DeleteCourse(ctx context.Context, courseID int) error {
	panic("not implemented")
}
