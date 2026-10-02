package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eaguilar88/deu/internal/certificates"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/eaguilar88/deu/internal/postgres_repository/queries"
)

func (r *PostgresRepository) GetCertificatesByCloseRequestID(ctx context.Context, closeRequestID int64) ([]entities.Certificate, error) {
	query, args, err := queries.GetCertificatesByCloseRequestID(closeRequestID).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entities.Certificate
	for rows.Next() {
		m, err := scanCertificate(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, newCertificateFromModel(m))
	}
	return result, rows.Err()
}

func (r *PostgresRepository) GetCertificateByVerificationCode(ctx context.Context, code string) (entities.Certificate, error) {
	query, args, err := queries.GetCertificateByVerificationCode(code).ToSql()
	if err != nil {
		return entities.Certificate{}, err
	}
	m, err := scanCertificate(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.Certificate{}, fmt.Errorf("%w: %w", certificates.ErrCertificateNotFound, err)
		}
		return entities.Certificate{}, err
	}
	return newCertificateFromModel(m), nil
}

func (r *PostgresRepository) MarkCertificateIssued(ctx context.Context, id int64, storageKey string) error {
	query, args, err := queries.MarkCertificateIssued(id, storageKey).ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func scanCertificate(row scannable) (models.Certificate, error) {
	var m models.Certificate
	err := row.Scan(
		&m.ID,
		&m.CloseRequestID,
		&m.CourseCycleID,
		&m.FirstName,
		&m.LastName,
		&m.Document,
		&m.Email,
		&m.VerificationCode,
		&m.StorageKey,
		&m.IssuedAt,
		&m.RevokedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	return m, err
}

func newCertificateFromModel(m models.Certificate) entities.Certificate {
	c := entities.Certificate{
		ID:               m.ID,
		CloseRequestID:   m.CloseRequestID,
		CourseCycleID:    m.CourseCycleID,
		FirstName:        m.FirstName,
		LastName:         m.LastName,
		Document:         m.Document,
		Email:            m.Email.String,
		VerificationCode: m.VerificationCode,
		StorageKey:       m.StorageKey.String,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	if m.IssuedAt.Valid {
		issuedAt := m.IssuedAt.Time
		c.IssuedAt = &issuedAt
	}
	if m.RevokedAt.Valid {
		revokedAt := m.RevokedAt.Time
		c.RevokedAt = &revokedAt
	}
	return c
}

func newCertificateModelFromEntity(c entities.Certificate) models.Certificate {
	return models.Certificate{
		CloseRequestID:   c.CloseRequestID,
		CourseCycleID:    c.CourseCycleID,
		FirstName:        c.FirstName,
		LastName:         c.LastName,
		Document:         c.Document,
		Email:            toNullString(c.Email),
		VerificationCode: c.VerificationCode,
	}
}
