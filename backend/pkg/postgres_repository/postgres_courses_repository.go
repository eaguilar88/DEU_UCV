package repository

import (
	"context"
	"database/sql"

	"github.com/eaguilar88/deu/pkg/entities"
	errs "github.com/eaguilar88/deu/pkg/errors"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
	"github.com/eaguilar88/deu/pkg/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetCourse(ctx context.Context, courseID string) (entities.Course, error) {
	query := queries.GetCourseByID(courseID)
	sql, args, err := query.ToSql()
	if err != nil {
		return entities.Course{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return entities.Course{}, err
	}
	defer stmt.Close()
	var course models.Course
	row := stmt.QueryRowContext(ctx, args...)
	course, err = scanCourse(row)
	if err != nil {
		return entities.Course{}, err
	}
	return newCourseFromModel(course), nil
}

func (r *PostgresRepository) GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	sql, args, err := queries.GetCourses(pageScope.PerPage, pageScope.Offset()).ToSql()
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
	var courses []entities.Course
	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		courses = append(courses, newCourseFromModel(course))
	}
	return courses, pageScope, nil
}

func (r *PostgresRepository) CreateCourse(ctx context.Context, course entities.Course) (int64, error) {
	sql, args, err := queries.InsertCourse(newCourseModelFromEntities(course)).ToSql()
	if err != nil {
		r.logger.Error("error creating query", zap.Error(err))
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		r.logger.Error("error preparing query", zap.Error(err))
		return -1, err
	}
	defer stmt.Close()
	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgErrorCodeUniqueViolation {
			r.logger.Error("error inserting user", zap.Error(err))
			return -1, errs.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting user", zap.Error(err))
		return -1, errs.NewInternalError(err)
	}
	return lastInsertedID, nil
}

func (r *PostgresRepository) UpdateCourse(ctx context.Context, courseID string, course entities.Course) error {
	sql, args, err := queries.UpdateCourse(courseID, newCourseModelFromEntities(course)).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errs.NewNotFoundError(err)
	}

	return nil
}

func (r *PostgresRepository) DeleteCourse(ctx context.Context, courseID string) error {
	sql, args, err := queries.DeleteCourse(courseID).ToSql()
	if err != nil {
		return errs.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, sql)
	if err != nil {
		return errs.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return errs.NewNotFoundError(err)
	}

	return nil
}

func scanCourse(row scannable) (models.Course, error) {
	var course models.Course
	err := row.Scan(
		&course.ID,
		&course.Name,
		&course.Description,
		&course.EndorsementID,
		&course.OwnerID,
		&course.OwnerFirstName,
		&course.OwnerLastName,
		&course.EndorserID,
		&course.EndorserFirstName,
		&course.EndorserLastName,
		&course.Content,
		&course.Objectives,
		&course.Cost,
		&course.Location,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	return course, err
}

func newCourseFromModel(course models.Course) entities.Course {
	var c = entities.Course{
		ID:   course.ID,
		Name: course.Name,
		Endorsement: entities.Endorsement{
			ID: course.EndorsementID,
			User: entities.User{
				ID:        course.EndorserID,
				FirstName: course.EndorserFirstName,
				LastName:  course.EndorserLastName,
			},
		},
		Owner: entities.User{
			ID:        course.OwnerID,
			FirstName: course.OwnerFirstName,
			LastName:  course.OwnerLastName,
		},
		Content:   course.Content,
		CreatedAt: course.CreatedAt,
		UpdatedAt: course.UpdatedAt,
	}

	if course.Description.Valid {
		c.Description = course.Description.String
	}

	if course.Objectives.Valid {
		c.Objectives = course.Objectives.String
	}

	if course.Cost.Valid {
		c.Cost = course.Cost.Float64
	}

	if course.Location.Valid {
		c.Location = course.Location.String
	}

	return c
}

func newCourseModelFromEntities(course entities.Course) models.Course {
	var c = models.Course{
		ID:                course.ID,
		Name:              course.Name,
		EndorsementID:     course.Endorsement.ID,
		OwnerID:           course.Owner.ID,
		OwnerFirstName:    course.Owner.FirstName,
		OwnerLastName:     course.Owner.LastName,
		EndorserID:        course.Endorsement.User.ID,
		EndorserFirstName: course.Endorsement.User.FirstName,
		EndorserLastName:  course.Endorsement.User.LastName,
		Description: sql.NullString{
			String: course.Description,
			Valid:  true,
		},
		Objectives: sql.NullString{
			String: course.Objectives,
			Valid:  true,
		},
		Cost: sql.NullFloat64{
			Float64: course.Cost,
			Valid:   true,
		},
		Location: sql.NullString{
			String: course.Location,
			Valid:  true,
		},
		Content:   course.Content,
		CreatedAt: course.CreatedAt,
		UpdatedAt: course.UpdatedAt,
	}

	return c
}
