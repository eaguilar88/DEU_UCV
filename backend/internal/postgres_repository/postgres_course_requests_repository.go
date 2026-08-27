package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eaguilar88/deu/internal/course_requests"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"go.uber.org/zap"
)

func (r *PostgresRepository) GetCourseRequestByID(ctx context.Context, requestID string) (entities.CourseRequest, error) {
	query, args, err := queries.GetCourseRequestByID(requestID).ToSql()
	if err != nil {
		return entities.CourseRequest{}, err
	}

	stmt, err := r.db.PrepareContext(ctx, query)
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
	found := false
	for rows.Next() {
		found = true
		request, err = scanCourseRequestWithCourse(rows)
		if err != nil {
			return entities.CourseRequest{}, err
		}
	}
	if !found {
		return entities.CourseRequest{}, fmt.Errorf("%w", course_requests.ErrCourseRequestNotFound)
	}
	return newCourseRequestFromModel(request), nil
}

func (r *PostgresRepository) GetCourseRequestsByFaculty(ctx context.Context, faculty entities.Faculty, scope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error) {
	query, args, err := queries.GetCourseRequestsByFaculty(faculty.String(), scope.PerPage, scope.Offset()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	r.logger.Debug("SQL Query: ", zap.String("query", query), zap.Any("args", args))
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
	var requests []entities.CourseRequest
	for rows.Next() {
		model, err := scanCourseRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		requests = append(requests, newCourseRequestFromModel(model))
	}
	var total int
	query, args, err = queries.CountCourseRequestsByFaculty(faculty.String()).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	if err = r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return nil, entities.PageScope{}, err
	}
	scope.Count = total
	return requests, scope, nil
}

func (r *PostgresRepository) GetCourseRequestsByProvider(ctx context.Context, providerID string, scope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error) {
	query, args, err := queries.GetCourseRequestsByProvider(providerID, scope.PerPage, scope.Offset()).ToSql()
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
	var requests []entities.CourseRequest
	for rows.Next() {
		model, err := scanCourseRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		requests = append(requests, newCourseRequestFromModel(model))
	}
	var total int
	query, args, err = queries.CountCourseRequestsByProvider(providerID).ToSql()
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	if err = r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return nil, entities.PageScope{}, err
	}
	scope.Count = total
	return requests, scope, nil
}

func (r *PostgresRepository) ApproveCourseRequest(ctx context.Context, reqID, reviewerID, courseType, comments string) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	// Get course ID from request
	var courseID string
	getCourseSQL := `SELECT course_id FROM deu.course_auth_requests WHERE id = $1`
	if err = tx.QueryRowContext(ctx, getCourseSQL, reqID).Scan(&courseID); err != nil {
		r.logger.Error("failed to get course ID from request", zap.Error(err))
		return err
	}

	// Update course type
	updateCourseSQL, updateCourseArgs, err := queries.UpdateCourseType(courseID, courseType, true).ToSql()
	if err != nil {
		r.logger.Error("failed to build update course type query", zap.Error(err))
		return err
	}
	if _, err = tx.ExecContext(ctx, updateCourseSQL, updateCourseArgs...); err != nil {
		r.logger.Error("failed to update course type", zap.Error(err))
		return err
	}

	// Approve request
	approveSQL, approveArgs, err := queries.ApproveCourseRequest(reqID, reviewerID, comments).ToSql()
	if err != nil {
		r.logger.Error("failed to build approve query", zap.Error(err))
		return err
	}
	res, err := tx.ExecContext(ctx, approveSQL, approveArgs...)
	if err != nil {
		r.logger.Error("failed to approve course request", zap.Error(err))
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		r.logger.Error("no rows affected when approving course request")
		return fmt.Errorf("%w", course_requests.ErrCourseRequestNotFound)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	r.logger.Info("course request approved successfully",
		zap.String("requestID", reqID),
		zap.String("courseID", courseID),
		zap.String("courseType", courseType),
	)
	return nil
}

func (r *PostgresRepository) RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string) error {
	query, args, err := queries.RejectCourseRequest(reqID, reviewerID, comments).ToSql()
	if err != nil {
		return err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", course_requests.ErrCourseRequestNotFound)
	}
	return nil
}

func (r *PostgresRepository) RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				r.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}
	}()

	// Get course ID from request
	getCourseSQL, getCourseArgs, err := queries.GetCourseIDFromRequest(reqID).ToSql()
	if err != nil {
		r.logger.Error("failed to build get course ID query", zap.Error(err))
		return err
	}
	var courseID string
	if err = tx.QueryRowContext(ctx, getCourseSQL, getCourseArgs...).Scan(&courseID); err != nil {
		r.logger.Error("failed to get course ID from request", zap.Error(err))
		return err
	}

	// Update course faculty (keep course inactive)
	updateCourseFacultySQL, updateCourseFacultyArgs, err := queries.UpdateCourseFaculty(courseID, faculty.String()).ToSql()
	if err != nil {
		r.logger.Error("failed to build update course faculty query", zap.Error(err))
		return err
	}
	if _, err = tx.ExecContext(ctx, updateCourseFacultySQL, updateCourseFacultyArgs...); err != nil {
		r.logger.Error("failed to update course faculty", zap.Error(err))
		return err
	}

	// Update request with reviewer and reason (comments)
	updateRequestSQL, updateRequestArgs, err := queries.RedirectCourseRequest(reqID, reviewerID, reason).ToSql()
	if err != nil {
		r.logger.Error("failed to build redirect request query", zap.Error(err))
		return err
	}
	res, err := tx.ExecContext(ctx, updateRequestSQL, updateRequestArgs...)
	if err != nil {
		r.logger.Error("failed to redirect course request", zap.Error(err))
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return err
	}
	if rowsAffected == 0 {
		r.logger.Error("no rows affected when redirecting course request")
		return fmt.Errorf("%w", course_requests.ErrCourseRequestNotFound)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	r.logger.Info("course request redirected successfully",
		zap.String("requestID", reqID),
		zap.String("courseID", courseID),
		zap.String("faculty", faculty.String()),
	)
	return nil
}

func (r *PostgresRepository) CreateCourseRequest(ctx context.Context, request entities.CourseRequest) (int64, error) {
	query, args, err := queries.InsertCourseRequest(newCourseRequestModelFromEntities(request)).ToSql()
	if err != nil {
		return -1, err
	}

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var lastInsertedID int64
	err = stmt.QueryRowContext(ctx, args...).Scan(&lastInsertedID)
	if err != nil {
		return -1, err
	}

	return lastInsertedID, nil
}

func scanCourseRequest(row scannable) (models.CourseRequest, error) {
	result := models.CourseRequest{}
	err := row.Scan(
		&result.ID,
		&result.CourseID,
		&result.Status,
		&result.ReviewerID,
		&result.Comments,
		&result.ReviewedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)

	return result, err
}

func scanCourseRequestWithCourse(row scannable) (models.CourseRequest, error) {
	result := models.CourseRequest{}
	course := models.Course{}
	err := row.Scan(
		// CourseRequest fields
		&result.ID,
		&result.CourseID,
		&result.Status,
		&result.ReviewerID,
		&result.Comments,
		&result.ReviewedAt,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
		// Course fields (excluding deleted_at)
		&course.ID,
		&course.Name,
		&course.Description,
		&course.OwnerID,
		&course.Objectives,
		&course.Duration,
		&course.Content,
		&course.Type,
		&course.Faculty,
		&course.Cost,
		&course.Location,
		&course.IsActive,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	if err != nil {
		return result, err
	}

	result.Course = &course
	return result, nil
}

func newCourseRequestFromModel(request models.CourseRequest) entities.CourseRequest {
	result := entities.CourseRequest{
		ID:        request.ID,
		Status:    entities.RequestStatus(request.Status),
		CreatedAt: request.CreatedAt,
		UpdatedAt: request.UpdatedAt,
	}

	if request.ReviewerID.Valid {
		result.Reviewer.ID = request.ReviewerID.String
	}

	if request.Comments.Valid {
		result.Comments = request.Comments.String
	}

	if request.ReviewedAt.Valid {
		result.ReviewedAt = request.ReviewedAt.String
	}

	// Map course if present
	if request.Course != nil {
		course := newCourseFromModel(*request.Course)
		result.Course = &course
	}

	return result
}

func newCourseRequestModelFromEntities(request entities.CourseRequest) models.CourseRequest {
	return models.CourseRequest{
		ID:       request.ID,
		CourseID: request.Course.ID,
		Status:   string(request.Status),
		ReviewerID: sql.NullString{
			String: request.Reviewer.ID,
			Valid:  request.Reviewer.ID != "",
		},
		Comments: sql.NullString{
			String: request.Comments,
			Valid:  request.Comments != "",
		},
		ReviewedAt: sql.NullString{
			String: request.ReviewedAt,
			Valid:  request.ReviewedAt != "",
		},
		CreatedAt: request.CreatedAt,
		UpdatedAt: request.UpdatedAt,
	}
}
