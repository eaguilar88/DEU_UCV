package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

var cycleCloseRequestSelectCommon = []string{
	"ccr.id",
	"ccr.course_cycle_id",
	"ccr.submitted_by",
	"ccr.status",
	"ccr.comments",
	"ccr.reviewer_id",
	"ccr.reviewed_at",
	"ccr.created_at",
	"ccr.updated_at",
}

func InsertCourseCycleCloseRequest(cycleID, submittedBy int64) sq.InsertBuilder {
	return psql.Insert(cycleCloseRequestsTableName).
		Columns("course_cycle_id", "submitted_by").
		Values(cycleID, submittedBy).
		Suffix("RETURNING id")
}

func CountPendingCloseRequestsForCycle(cycleID int64) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(cycleCloseRequestsTableName).
		Where(sq.Eq{"course_cycle_id": cycleID}).
		Where(sq.Eq{"status": "under_review"}).
		Where(sq.Eq{"deleted_at": nil})
}

func GetCourseCycleCloseRequestByID(id string) sq.SelectBuilder {
	return psql.Select(cycleCloseRequestSelectCommon...).
		From(fmt.Sprintf("%s AS ccr", cycleCloseRequestsTableName)).
		Where(sq.Eq{"ccr.id": id}).
		Where(sq.Eq{"ccr.deleted_at": nil})
}

func GetCourseCycleCloseRequests(faculty string, perPage, offset uint64) sq.SelectBuilder {
	q := psql.Select(cycleCloseRequestSelectCommon...).
		From(fmt.Sprintf("%s AS ccr", cycleCloseRequestsTableName)).
		Where(sq.Eq{"ccr.deleted_at": nil})

	if faculty != "" {
		q = q.Join(fmt.Sprintf("%s AS pd ON pd.id = ccr.course_cycle_id", periodsTableName)).
			Join(fmt.Sprintf("%s AS c ON c.id = pd.course_id", coursesTableName)).
			Where(sq.Eq{"c.faculty": faculty})
	}

	return q.OrderBy("ccr.created_at DESC").
		Limit(perPage).
		Offset(offset)
}

func CountCourseCycleCloseRequests() sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(cycleCloseRequestsTableName).
		Where(sq.Eq{"deleted_at": nil})
}

func ApproveCourseCycleCloseRequest(id, reviewerID string) sq.UpdateBuilder {
	return psql.Update(cycleCloseRequestsTableName).
		Set("status", "approved").
		Set("reviewer_id", reviewerID).
		Set("reviewed_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}

func RejectCourseCycleCloseRequest(id, reviewerID, comments string) sq.UpdateBuilder {
	return psql.Update(cycleCloseRequestsTableName).
		Set("status", "rejected").
		Set("reviewer_id", reviewerID).
		Set("comments", comments).
		Set("reviewed_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}

// SetCourseCycleClosed closes every currently-active cycle belonging to the same course as
// cycleID (not just cycleID itself) — this is the cascade Node performs on close-request
// approval. Today at most one row is ever active per course (see the §2.1 guard in
// course_periods), but this makes that cascade explicit and correct-by-construction.
func SetCourseCycleClosed(cycleID string) sq.UpdateBuilder {
	return psql.Update(periodsTableName).
		Set("closed_at", sq.Expr("NOW()")).
		Set("is_active", false).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Expr(fmt.Sprintf("course_id = (SELECT course_id FROM %s WHERE id = ?)", periodsTableName), cycleID)).
		Where(sq.Eq{"is_active": true})
}

// SetCourseManagementStatusByCycleID sets (or, with status == nil, clears) the estado_gestion
// of the course that owns the period referenced by cycleID.
func SetCourseManagementStatusByCycleID(cycleID string, status *string) sq.UpdateBuilder {
	return psql.Update(coursesTableName).
		Set("estado_gestion", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Expr(fmt.Sprintf("id = (SELECT course_id FROM %s WHERE id = ?)", periodsTableName), cycleID))
}
