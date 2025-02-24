package repository

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
	"github.com/eaguilar88/deu/pkg/repository/queries"
)

func (r *PostgresRepository) GetCourse(ctx context.Context, courseID int) (entities.Course, error) {
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
	sql, args, err := queries.GetCourses(pageScope).ToSql()
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
	panic("not implemented")
}

func (r *PostgresRepository) UpdateCourse(ctx context.Context, courseID int, course entities.Course) error {
	panic("not implemented")
}

func (r *PostgresRepository) DeleteCourse(ctx context.Context, courseID int) error {
	panic("not implemented")
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
		Endorsement: entities.Endorsements{
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
