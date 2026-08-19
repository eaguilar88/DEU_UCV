package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var resourceRequestSelectCommon = []string{
	"r.id",
	"r.group_id",
	"COALESCE(g.name, '') AS group_name",
	"r.type",
	"r.content",
	"r.status",
	"r.created_at",
	"r.updated_at",
}

func GetGroupResourceRequestByID(reqID string) sq.SelectBuilder {
	return psql.Select(resourceRequestSelectCommon...).
		From(fmt.Sprintf("%s AS r", groupResourceRequestsTableName)).
		LeftJoin(fmt.Sprintf("%s AS g ON r.group_id = g.id", groupsTableName)).
		Where(sq.Eq{"r.id": reqID, "r.deleted_at": nil})
}

func GetGroupResourceRequestsByFaculty(faculty string, status string, limit, offset int) sq.SelectBuilder {
	q := psql.Select(resourceRequestSelectCommon...).
		From(fmt.Sprintf("%s AS r", groupResourceRequestsTableName)).
		Join(fmt.Sprintf("%s AS g ON r.group_id = g.id", groupsTableName)).
		Where(sq.Eq{"g.faculty": faculty, "r.deleted_at": nil})

	if status != "" {
		q = q.Where(sq.Eq{"r.status": status})
	}

	return q.OrderBy("r.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func CountGroupResourceRequestsByFaculty(faculty string, status string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(fmt.Sprintf("%s AS r", groupResourceRequestsTableName)).
		Join(fmt.Sprintf("%s AS g ON r.group_id = g.id", groupsTableName)).
		Where(sq.Eq{"g.faculty": faculty, "r.deleted_at": nil})
}

func GetGroupResourceRequestsByGroupID(groupID string, status string, limit, offset int) sq.SelectBuilder {
	q := psql.Select(resourceRequestSelectCommon...).
		From(fmt.Sprintf("%s AS r", groupResourceRequestsTableName)).
		LeftJoin(fmt.Sprintf("%s AS g ON r.group_id = g.id", groupsTableName)).
		Where(sq.Eq{"r.group_id": groupID, "r.deleted_at": nil})

	if status != "" {
		q = q.Where(sq.Eq{"r.status": status})
	}

	return q.OrderBy("r.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func CountGroupResourceRequestsByGroupID(groupID string, status string) sq.SelectBuilder {
	q := psql.Select("COUNT(*)").
		From(fmt.Sprintf("%s AS r", groupResourceRequestsTableName)).
		Where(sq.Eq{"r.group_id": groupID, "r.deleted_at": nil})

	if status != "" {
		q = q.Where(sq.Eq{"r.status": status})
	}

	return q
}

func CountPendingGroupResourceRequestsByFaculty(faculty string) sq.SelectBuilder {
	q := psql.Select("g.faculty", "COUNT(r.id) AS pending_count").
		From(fmt.Sprintf("%s AS r", groupResourceRequestsTableName)).
		Join(fmt.Sprintf("%s AS g ON r.group_id = g.id", groupsTableName)).
		Where(sq.Eq{"r.status": "under_review", "r.deleted_at": nil})

	if faculty != "" {
		q = q.Where(sq.Eq{"g.faculty": faculty})
	}

	return q.GroupBy("g.faculty")
}

func InsertGroupResourceRequest(m models.GroupResourceRequest) sq.InsertBuilder {
	return psql.Insert(groupResourceRequestsTableName).
		Columns("group_id", "type", "content", "status", "created_at", "updated_at").
		Values(m.GroupID, m.Type, m.Content, m.Status, sq.Expr("NOW()"), sq.Expr("NOW()")).
		Suffix("RETURNING id")
}

func ApproveGroupResourceRequest(reqID string) sq.UpdateBuilder {
	return psql.Update(groupResourceRequestsTableName).
		Set("status", "approved").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": reqID, "status": "under_review", "deleted_at": nil})
}

func RejectGroupResourceRequest(reqID string) sq.UpdateBuilder {
	return psql.Update(groupResourceRequestsTableName).
		Set("status", "rejected").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": reqID, "status": "under_review", "deleted_at": nil})
}
