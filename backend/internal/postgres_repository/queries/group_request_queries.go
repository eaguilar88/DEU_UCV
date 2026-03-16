package queries

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var groupAuthRequestQuerySelectCommon = []string{
	"id",
	"group_id",
	"faculty",
	"status",
	"comments",
	"created_at",
	"updated_at",
	"reviewer_id",
	"reviewed_at",
}

func GetGroupRequestByID(requestID string) sq.SelectBuilder {
	return psql.Select(
		groupAuthRequestQuerySelectCommon...,
	).From(groupRequestsTableName).Where(sq.Eq{"id": requestID})
}

func GetGroupRequestsByFaculty(faculty string, limit, offset int) sq.SelectBuilder {
	return psql.Select(
		groupAuthRequestQuerySelectCommon...,
	).From(groupRequestsTableName).Where(sq.Eq{"faculty": faculty}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func ApproveGroupRequest(requestID string) sq.UpdateBuilder {
	return psql.Update(groupRequestsTableName).
		Set("status", "approved").
		Set("reviewed_at", "NOW()").
		Where(sq.Eq{"id": requestID, "status": "under_review"})
}

func RejectGroupRequest(requestID string) sq.UpdateBuilder {
	return psql.Update(groupRequestsTableName).
		Set("status", "rejected").
		Set("reviewed_at", "NOW()").
		Where(sq.Eq{"id": requestID, "status": "under_review"})
}

func InsertGroupRequest(req models.GroupRequest) sq.InsertBuilder {
	return psql.Insert(groupRequestsTableName).
		Columns(
			"group_id",
			"faculty",
			"status",
			"comments",
			"created_at",
			"updated_at",
		).
		Values(
			req.GroupID,
			req.Faculty,
			req.Status,
			req.Comments,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func CountGroupRequestsByFaculty(faculty string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").From(groupRequestsTableName).Where(sq.Eq{"faculty": faculty})
}
