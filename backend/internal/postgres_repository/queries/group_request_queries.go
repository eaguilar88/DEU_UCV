package queries

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var groupAuthRequestQuerySelectCommon = []string{
	"gr.id",
	"gr.group_id",
	"g.name AS group_name",
	"gr.faculty",
	"gr.status",
	"gr.comments",
	"gr.created_at",
	"gr.updated_at",
	"gr.reviewer_id",
	"gr.reviewed_at",
}

func GetGroupRequestByID(requestID string) sq.SelectBuilder {
	return psql.Select(
		groupAuthRequestQuerySelectCommon...,
	).From(groupRequestsTableName + " gr").
		InnerJoin("deu.extension_groups g ON gr.group_id = g.id").
		Where(sq.Eq{"gr.id": requestID})
}

func GetGroupRequestsByFaculty(faculty string, status string, limit, offset int) sq.SelectBuilder {
	q := psql.Select(
		groupAuthRequestQuerySelectCommon...,
	).From(groupRequestsTableName + " gr").
		InnerJoin("deu.extension_groups g ON gr.group_id = g.id").
		Where(sq.Eq{"gr.faculty": faculty})

	if status != "" {
		q = q.Where(sq.Eq{"gr.status": status})
	}

	return q.OrderBy("gr.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func CountGroupRequestsByFaculty(faculty string, status string) sq.SelectBuilder {
	q := psql.Select("COUNT(*)").From(groupRequestsTableName + " gr").Where(sq.Eq{"gr.faculty": faculty})
	if status != "" {
		q = q.Where(sq.Eq{"gr.status": status})
	}
	return q
}

func CountPendingGroupRequestsByFaculty(faculty string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(groupRequestsTableName + " gr").
		Where(sq.Eq{"gr.faculty": faculty, "gr.status": "under_review"})
}

func GetPendingGroupRequestsCountGroupedByFaculty(faculty string) sq.SelectBuilder {
	q := psql.Select("faculty", "COUNT(*) as count").
		From(groupRequestsTableName).
		Where(sq.Eq{"status": "under_review"})

	if faculty != "" {
		q = q.Where(sq.Eq{"faculty": faculty})
	}

	return q.GroupBy("faculty")
}

func GetGroupRequestsByGroupID(groupID string) sq.SelectBuilder {
	return psql.Select(
		groupAuthRequestQuerySelectCommon...,
	).From(groupRequestsTableName + " gr").
		InnerJoin("deu.extension_groups g ON gr.group_id = g.id").
		Where(sq.Eq{"gr.group_id": groupID}).
		OrderBy("gr.created_at DESC")
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

func ActivateGroup(groupID string) sq.UpdateBuilder {
	return psql.Update("deu.extension_groups").
		Set("is_active", true).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": groupID})
}
