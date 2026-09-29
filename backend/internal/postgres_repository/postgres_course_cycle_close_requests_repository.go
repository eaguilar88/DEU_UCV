package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/internal/course_cycle_close_requests"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
	"go.uber.org/zap"
)

func (r *PostgresRepository) HasPendingCloseRequestForCycle(ctx context.Context, cycleID int64) (bool, error) {
	query, args, err := queries.CountPendingCloseRequestsForCycle(cycleID).ToSql()
	if err != nil {
		return false, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var count int
	if err := stmt.QueryRowContext(ctx, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateCourseCycleCloseRequest creates the close request together with one pending (not yet
// issued) certificate per approved participant.
func (r *PostgresRepository) CreateCourseCycleCloseRequest(ctx context.Context, cycleID, submittedByID int64, certificates []entities.Certificate) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	insertQuery, insertArgs, err := queries.InsertCourseCycleCloseRequest(cycleID, submittedByID).ToSql()
	if err != nil {
		r.logger.Error("error creating cycle close request query", zap.Error(err))
		return -1, err
	}
	var id int64
	if err = tx.QueryRowContext(ctx, insertQuery, insertArgs...).Scan(&id); err != nil {
		r.logger.Error("error inserting cycle close request", zap.Error(err))
		return -1, err
	}

	if len(certificates) > 0 {
		certModels := make([]models.Certificate, 0, len(certificates))
		for _, c := range certificates {
			c.CloseRequestID = id
			c.CourseCycleID = cycleID
			certModels = append(certModels, newCertificateModelFromEntity(c))
		}
		certQuery, certArgs, qerr := queries.InsertCertificates(certModels).ToSql()
		if qerr != nil {
			err = qerr
			return -1, httperrors.NewBadQueryError(err)
		}
		if _, err = tx.ExecContext(ctx, certQuery, certArgs...); err != nil {
			r.logger.Error("error inserting close request certificates", zap.Error(err))
			return -1, err
		}
	}

	closureRequested := string(entities.CourseManagementStatusClosureRequested)
	statusQuery, statusArgs, err := queries.SetCourseManagementStatusByCycleID(strconv.FormatInt(cycleID, 10), &closureRequested).ToSql()
	if err != nil {
		return -1, httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, statusQuery, statusArgs...); err != nil {
		r.logger.Error("error setting course management status", zap.Error(err))
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}
	return id, nil
}

func (r *PostgresRepository) GetCourseCycleCloseRequestByID(ctx context.Context, id string) (entities.CourseCycleCloseRequest, error) {
	query, args, err := queries.GetCourseCycleCloseRequestByID(id).ToSql()
	if err != nil {
		return entities.CourseCycleCloseRequest{}, err
	}
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return entities.CourseCycleCloseRequest{}, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, args...)
	m, err := scanCourseCycleCloseRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.CourseCycleCloseRequest{}, fmt.Errorf("%w: %w", course_cycle_close_requests.ErrCycleCloseRequestNotFound, err)
		}
		return entities.CourseCycleCloseRequest{}, err
	}
	return newCourseCycleCloseRequestFromModel(m), nil
}

func (r *PostgresRepository) GetCourseCycleCloseRequestByCertificatesToken(ctx context.Context, token string) (entities.CourseCycleCloseRequest, error) {
	query, args, err := queries.GetCourseCycleCloseRequestByCertificatesToken(token).ToSql()
	if err != nil {
		return entities.CourseCycleCloseRequest{}, err
	}
	m, err := scanCourseCycleCloseRequest(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.CourseCycleCloseRequest{}, fmt.Errorf("%w: %w", course_cycle_close_requests.ErrCycleCloseRequestNotFound, err)
		}
		return entities.CourseCycleCloseRequest{}, err
	}
	return newCourseCycleCloseRequestFromModel(m), nil
}

func (r *PostgresRepository) GetCourseCycleCloseRequests(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseCycleCloseRequest, entities.PageScope, error) {
	query, args, err := queries.GetCourseCycleCloseRequests(faculty.String(), uint64(pageScope.PerPage), uint64(pageScope.Offset())).ToSql()
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
	var requests []entities.CourseCycleCloseRequest
	for rows.Next() {
		m, err := scanCourseCycleCloseRequest(rows)
		if err != nil {
			return nil, entities.PageScope{}, err
		}
		requests = append(requests, newCourseCycleCloseRequestFromModel(m))
	}
	pageScope.Count = len(requests)
	return requests, pageScope, nil
}

// ApproveCourseCycleCloseRequest approves the request, closes its cycle and enqueues certificatesJob,
// all in one transaction, so the certificates are generated only if the approval commits.
func (r *PostgresRepository) ApproveCourseCycleCloseRequest(ctx context.Context, id, reviewerID, cycleID, certificatesToken string, certificatesJob entities.Job, notify func() error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	approveQuery, approveArgs, err := queries.ApproveCourseCycleCloseRequest(id, reviewerID, certificatesToken).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, approveQuery, approveArgs...); err != nil {
		r.logger.Error("error approving cycle close request", zap.Error(err))
		return err
	}

	closeQuery, closeArgs, err := queries.SetCourseCycleClosed(cycleID).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, closeQuery, closeArgs...); err != nil {
		r.logger.Error("error closing course cycle", zap.Error(err))
		return err
	}

	closed := string(entities.CourseManagementStatusClosed)
	statusQuery, statusArgs, err := queries.SetCourseManagementStatusByCycleID(cycleID, &closed).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, statusQuery, statusArgs...); err != nil {
		r.logger.Error("error setting course management status", zap.Error(err))
		return err
	}

	jobQuery, jobArgs, err := queries.InsertJob(certificatesJob.Kind, certificatesJob.Payload).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, jobQuery, jobArgs...); err != nil {
		r.logger.Error("error enqueuing certificates job", zap.Error(err))
		return err
	}

	if err = notify(); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) RejectCourseCycleCloseRequest(ctx context.Context, id, reviewerID, comments, cycleID string, notify func() error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	rejectQuery, rejectArgs, err := queries.RejectCourseCycleCloseRequest(id, reviewerID, comments).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	result, err := tx.ExecContext(ctx, rejectQuery, rejectArgs...)
	if err != nil {
		return err
	}
	if affected, rerr := result.RowsAffected(); rerr != nil || affected == 0 {
		err = fmt.Errorf("%w: %w", course_cycle_close_requests.ErrCycleCloseRequestNotFound, rerr)
		return err
	}

	open := string(entities.CourseManagementStatusOpen)
	statusQuery, statusArgs, err := queries.SetCourseManagementStatusByCycleID(cycleID, &open).ToSql()
	if err != nil {
		return httperrors.NewBadQueryError(err)
	}
	if _, err = tx.ExecContext(ctx, statusQuery, statusArgs...); err != nil {
		r.logger.Error("error setting course management status", zap.Error(err))
		return err
	}

	if err = notify(); err != nil {
		return err
	}

	return tx.Commit()
}

func scanCourseCycleCloseRequest(row scannable) (models.CourseCycleCloseRequest, error) {
	var m models.CourseCycleCloseRequest
	err := row.Scan(
		&m.ID,
		&m.CourseCycleID,
		&m.SubmittedBy,
		&m.Status,
		&m.Comments,
		&m.ReviewerID,
		&m.ReviewedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.CertificatesToken,
	)
	return m, err
}

func newCourseCycleCloseRequestFromModel(m models.CourseCycleCloseRequest) entities.CourseCycleCloseRequest {
	req := entities.CourseCycleCloseRequest{
		ID:            m.ID,
		CourseCycleID: m.CourseCycleID,
		SubmittedByID: strconv.FormatInt(m.SubmittedBy, 10),
		Status:        entities.RequestStatus(m.Status),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.Comments.Valid {
		req.Comments = m.Comments.String
	}
	if m.ReviewedAt.Valid {
		req.ReviewedAt = m.ReviewedAt.String
	}
	if m.CertificatesToken.Valid {
		req.CertificatesToken = m.CertificatesToken.String
	}
	return req
}
