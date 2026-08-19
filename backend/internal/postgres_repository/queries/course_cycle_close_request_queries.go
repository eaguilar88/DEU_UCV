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

func SetCourseCycleClosed(cycleID string) sq.UpdateBuilder {
	return psql.Update(periodsTableName).
		Set("closed_at", sq.Expr("NOW()")).
		Set("is_active", false).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": cycleID})
}
