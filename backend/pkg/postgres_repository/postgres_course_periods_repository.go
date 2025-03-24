package repository

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
	errs "github.com/eaguilar88/deu/pkg/errors"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
	"github.com/eaguilar88/deu/pkg/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
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

func (r *PostgresRepository) GetCoursePeriods(ctx context.Context, courseID int, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
	sql, args, err := queries.GetCoursePeriods(courseID, pageScope).ToSql()
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
	var coursePeriods []entities.CoursePeriod
	for rows.Next() {
		coursePeriod, err := scanCoursePeriod(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		coursePeriods = append(coursePeriods, newCoursePeriodFromModel(coursePeriod))
	}
	return coursePeriods, pageScope, nil
}

func (r *PostgresRepository) CreateCoursePeriod(ctx context.Context, coursePeriod entities.CoursePeriod) (int64, error) {
	cpModel := models.CoursePeriod{
		ID:              coursePeriod.ID,
		CourseID:        coursePeriod.Course.ID,
		StartDate:       coursePeriod.StartDate,
		EndDate:         coursePeriod.EndDate,
		InscriptionDate: coursePeriod.InscriptionDate,
	}

	sql, args, err := queries.InsertCoursePeriod(cpModel).ToSql()
	if err != nil {
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("duplicated course period", zap.Error(err))
			return -1, errs.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting course period", zap.Error(err))
		return -1, err
	}
	return lastInsertedID, nil
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
		&coursePeriod.CourseID,
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
	var deletedAt string

	if coursePeriod.DeletedAt.Valid {
		deletedAt = coursePeriod.DeletedAt.String
	}

	return entities.CoursePeriod{
		ID: coursePeriod.ID,
		Course: entities.Course{
			ID: coursePeriod.CourseID,
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
