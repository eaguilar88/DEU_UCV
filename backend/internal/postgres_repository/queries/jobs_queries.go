package queries

import (
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"
)

var jobReturningColumns = "RETURNING id, kind, payload, status, attempts, max_attempts, run_at, last_error, created_at, updated_at"

func InsertJob(kind string, payload []byte) sq.InsertBuilder {
	return psql.Insert(jobsTableName).
		Columns("kind", "payload").
		Values(kind, string(payload)).
		Suffix("RETURNING id")
}

// ClaimJob marks the next runnable job as running and leases it for leaseSeconds. A job is
// runnable when it is pending and due, or when it is running but its lease expired (the worker
// that held it died). SKIP LOCKED keeps concurrent workers from claiming the same job.
func ClaimJob(leaseSeconds int) sq.UpdateBuilder {
	return psql.Update(jobsTableName).
		Set("status", "running").
		Set("attempts", sq.Expr("attempts + 1")).
		Set("locked_until", sq.Expr("NOW() + (? * INTERVAL '1 second')", leaseSeconds)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Expr(fmt.Sprintf(`id = (
			SELECT id FROM %s
			WHERE attempts < max_attempts
			  AND ((status = 'pending' AND run_at <= NOW()) OR (status = 'running' AND locked_until < NOW()))
			ORDER BY run_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)`, jobsTableName))).
		Suffix(jobReturningColumns)
}

func CompleteJob(id int64) sq.UpdateBuilder {
	return psql.Update(jobsTableName).
		Set("status", "done").
		Set("locked_until", nil).
		Set("last_error", nil).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}

// RetryJobLater puts a failed attempt back in the queue to run again after delaySeconds.
func RetryJobLater(id int64, lastError string, delaySeconds int) sq.UpdateBuilder {
	return psql.Update(jobsTableName).
		Set("status", "pending").
		Set("run_at", sq.Expr("NOW() + (? * INTERVAL '1 second')", delaySeconds)).
		Set("locked_until", nil).
		Set("last_error", lastError).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}

func FailJob(id int64, lastError string) sq.UpdateBuilder {
	return psql.Update(jobsTableName).
		Set("status", "failed").
		Set("locked_until", nil).
		Set("last_error", lastError).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}

// ResetFailedCourseCycleCertificatesJob requeues the failed certificates job of a close request
// with a fresh set of attempts.
func ResetFailedCourseCycleCertificatesJob(kind string, closeRequestID int64) sq.UpdateBuilder {
	return psql.Update(jobsTableName).
		Set("status", "pending").
		Set("attempts", 0).
		Set("run_at", sq.Expr("NOW()")).
		Set("locked_until", nil).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"kind": kind}).
		Where(sq.Eq{"status": "failed"}).
		Where(sq.Expr("payload->>'close_request_id' = ?", strconv.FormatInt(closeRequestID, 10)))
}
