package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/internal/courses"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetCourse(ctx context.Context, courseID string) (entities.Course, error) {
	query, args, err := queries.GetCourseByID(courseID).ToSql()
	if err != nil {
		return entities.Course{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.Course{}, err
	}
	defer stmt.Close()
	var course models.Course
	row := stmt.QueryRowContext(ctx, args...)
	course, err = scanCourse(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.Course{}, fmt.Errorf("%w: %w", courses.ErrCourseNotFound, err)
		}
		return entities.Course{}, err
	}
	return newCourseFromModel(course), nil
}

func (r *PostgresRepository) GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	query, args, err := queries.GetCourses(pageScope.PerPage, pageScope.Offset()).ToSql()
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
	query, args, err := queries.InsertCourse(newCourseModelFromEntities(course)).ToSql()
	if err != nil {
		r.logger.Error("error creating query", zap.Error(err))
		return -1, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
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
			return -1, httperrors.NewDuplicateEntryError(err)
		}
		r.logger.Error("error inserting user", zap.Error(err))
		return -1, httperrors.NewInternal(err)
	}
	return lastInsertedID, nil
}

// CreateCourseWithRequest creates a course and its authorization request in a single transaction
func (r *PostgresRepository) CreateCourseWithRequest(ctx context.Context, course entities.Course) (courseID int64, requestID int64, err error) {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	courseSQL, courseArgs, err := queries.InsertCourse(newCourseModelFromEntities(course)).ToSql()
	if err != nil {
		r.logger.Error("failed to build course query", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to build course query: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, courseSQL)
	if err != nil {
		r.logger.Error("failed to prepare course statement", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to prepare course statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, courseArgs...).Scan(&courseID)
	if err != nil {
		r.logger.Error("failed to insert course", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to insert course: %w", err)
	}

	// 2. Create the course authorization request
	courseRequest := entities.CourseRequest{
		Course: &entities.Course{
			ID: strconv.FormatInt(courseID, 10),
		},
		Status: entities.RequestStatus_UNDER_REVIEW,
	}

	requestSQL, requestArgs, err := queries.InsertCourseRequest(newCourseRequestModelFromEntities(courseRequest)).ToSql()
	if err != nil {
		r.logger.Error("failed to build request query", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to build request query: %w", err)
	}

	stmt2, err := tx.PrepareContext(ctx, requestSQL)
	if err != nil {
		r.logger.Error("failed to prepare request statement", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to prepare request statement: %w", err)
	}
	defer stmt2.Close()

	err = stmt2.QueryRowContext(ctx, requestArgs...).Scan(&requestID)
	if err != nil {
		r.logger.Error("failed to insert course request", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to insert course request: %w", err)
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return -1, -1, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("course and request created successfully",
		zap.Int64("course_id", courseID),
		zap.Int64("request_id", requestID))

	return courseID, requestID, nil
}

func (r *PostgresRepository) UpdateCourse(ctx context.Context, courseID string, course entities.Course) error {
	query, args, err := queries.UpdateCourse(courseID, newCourseModelFromEntities(course)).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w: %w", courses.ErrCourseNotFound, err)
	}

	return nil
}

func (r *PostgresRepository) DeleteCourse(ctx context.Context, courseID string) error {
	query, args, err := queries.DeleteCourse(courseID).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		return fmt.Errorf("%w: %w", courses.ErrCourseNotFound, err)
	}

	return nil
}

func scanCourse(row scannable) (models.Course, error) {
	var course models.Course
	err := row.Scan(
		&course.ID,
		&course.Name,
		&course.Description,
		&course.OwnerID,
		&course.Objectives,
		&course.Rationale,
		&course.Duration,
		&course.Cost,
		&course.InstructorProfile,
		&course.Profiles,
		&course.Requirements,
		&course.Content,
		&course.Evaluation,
		&course.Schedule,
		&course.Type,
		&course.Faculty,
		&course.Location,
		&course.IsActive,
		&course.HasDocumentation,
		&course.CreatedAt,
		&course.UpdatedAt,
		&course.DeletedAt,
	)

	return course, err
}

func newCourseFromModel(course models.Course) entities.Course {
	c := entities.Course{
		ID:               course.ID,
		Name:             course.Name,
		HasDocumentation: course.HasDocumentation,
		CreatedAt:        course.CreatedAt,
		UpdatedAt:        course.UpdatedAt,
	}

	if course.Description.Valid {
		c.Description = course.Description.String
	}

	if course.Objectives.Valid {
		c.Objectives = course.Objectives.String
	}

	if course.Rationale.Valid {
		c.Rationale = course.Rationale.String
	}

	if course.Duration.Valid {
		c.Duration = course.Duration.String
	}

	if course.Cost.Valid {
		c.Cost = course.Cost.String
	}

	if course.InstructorProfile.Valid {
		c.InstructorProfile = course.InstructorProfile.String
	}

	if course.Profiles.Valid {
		c.Profiles = course.Profiles.String
	}

	if course.Requirements.Valid {
		c.Requirements = course.Requirements.String
	}

	if course.Content.Valid {
		c.Content = course.Content.String
	}

	if course.Evaluation.Valid {
		c.Evaluation = course.Evaluation.String
	}

	if course.Schedule.Valid {
		c.Schedule = course.Schedule.String
	}

	if course.Type.Valid {
		c.Type = entities.FromStringCourseType(course.Type.String)
	}

	if course.Faculty.Valid {
		faculty, err := entities.FromString(course.Faculty.String)
		if err == nil {
			c.Faculty = faculty
		}
	}

	if course.Location.Valid {
		c.Location = course.Location.String
	}

	return c
}

func newCourseModelFromEntities(course entities.Course) models.Course {
	return models.Course{
		ID:      course.ID,
		Name:    course.Name,
		OwnerID: course.Owner.ID,
		Description: sql.NullString{
			String: course.Description,
			Valid:  course.Description != "",
		},
		Objectives: sql.NullString{
			String: course.Objectives,
			Valid:  course.Objectives != "",
		},
		Rationale: sql.NullString{
			String: course.Rationale,
			Valid:  course.Rationale != "",
		},
		Duration: sql.NullString{
			String: course.Duration,
			Valid:  course.Duration != "",
		},
		Cost: sql.NullString{
			String: course.Cost,
			Valid:  course.Cost != "",
		},
		InstructorProfile: sql.NullString{
			String: course.InstructorProfile,
			Valid:  course.InstructorProfile != "",
		},
		Profiles: sql.NullString{
			String: course.Profiles,
			Valid:  course.Profiles != "",
		},
		Requirements: sql.NullString{
			String: course.Requirements,
			Valid:  course.Requirements != "",
		},
		Content: sql.NullString{
			String: course.Content,
			Valid:  course.Content != "",
		},
		Evaluation: sql.NullString{
			String: course.Evaluation,
			Valid:  course.Evaluation != "",
		},
		Schedule: sql.NullString{
			String: course.Schedule,
			Valid:  course.Schedule != "",
		},
		Type: sql.NullString{
			String: string(course.Type),
			Valid:  course.Type != "",
		},
		Faculty: sql.NullString{
			String: string(course.Faculty),
			Valid:  course.Faculty != "",
		},
		Location: sql.NullString{
			String: course.Location,
			Valid:  course.Location != "",
		},
		IsActive:         false,
		HasDocumentation: course.HasDocumentation,
		CreatedAt:        course.CreatedAt,
		UpdatedAt:        course.UpdatedAt,
	}
}
