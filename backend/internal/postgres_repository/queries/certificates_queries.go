package queries

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var certificateSelectCommon = []string{
	"id",
	"close_request_id",
	"course_cycle_id",
	"first_name",
	"last_name",
	"document",
	"email",
	"verification_code",
	"storage_key",
	"issued_at",
	"revoked_at",
	"created_at",
	"updated_at",
}

func InsertCertificates(certificates []models.Certificate) sq.InsertBuilder {
	q := psql.Insert(certificatesTableName).
		Columns("close_request_id", "course_cycle_id", "first_name", "last_name", "document", "email", "verification_code")
	for _, c := range certificates {
		q = q.Values(c.CloseRequestID, c.CourseCycleID, c.FirstName, c.LastName, c.Document, c.Email, c.VerificationCode)
	}
	return q
}

func GetCertificatesByCloseRequestID(closeRequestID int64) sq.SelectBuilder {
	return psql.Select(certificateSelectCommon...).
		From(certificatesTableName).
		Where(sq.Eq{"close_request_id": closeRequestID}).
		OrderBy("last_name", "first_name", "id")
}

func GetCertificateByVerificationCode(code string) sq.SelectBuilder {
	return psql.Select(certificateSelectCommon...).
		From(certificatesTableName).
		Where(sq.Eq{"verification_code": code})
}

func MarkCertificateIssued(id int64, storageKey string) sq.UpdateBuilder {
	return psql.Update(certificatesTableName).
		Set("storage_key", storageKey).
		Set("issued_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}
