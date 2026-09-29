package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/certificates"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

// ClaimJob leases the next runnable job for lease. It returns nil when there is nothing to run.
func (r *PostgresRepository) ClaimJob(ctx context.Context, lease time.Duration) (*entities.Job, error) {
	query, args, err := queries.ClaimJob(int(lease.Seconds())).ToSql()
	if err != nil {
		return nil, err
	}
	var m models.Job
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&m.ID,
		&m.Kind,
		&m.Payload,
		&m.Status,
		&m.Attempts,
		&m.MaxAttempts,
		&m.RunAt,
		&m.LastError,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	job := newJobFromModel(m)
	return &job, nil
}

func (r *PostgresRepository) CompleteJob(ctx context.Context, id int64) error {
	query, args, err := queries.CompleteJob(id).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresRepository) RetryJobLater(ctx context.Context, id int64, lastError string, delay time.Duration) error {
	query, args, err := queries.RetryJobLater(id, lastError, int(delay.Seconds())).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresRepository) FailJob(ctx context.Context, id int64, lastError string) error {
	query, args, err := queries.FailJob(id, lastError).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresRepository) ResetFailedCertificatesJob(ctx context.Context, closeRequestID int64) error {
	query, args, err := queries.ResetFailedCourseCycleCertificatesJob(entities.JobKindCourseCycleCertificates, closeRequestID).ToSql()
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w: close request %d", certificates.ErrNoFailedCertificatesJob, closeRequestID)
	}
	return nil
}

func newJobFromModel(m models.Job) entities.Job {
	return entities.Job{
		ID:          m.ID,
		Kind:        m.Kind,
		Payload:     m.Payload,
		Status:      entities.JobStatus(m.Status),
		Attempts:    m.Attempts,
		MaxAttempts: m.MaxAttempts,
		RunAt:       m.RunAt,
		LastError:   m.LastError.String,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
