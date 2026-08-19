package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

var providerRequestSelectCommon = []string{
	"pr.id",
	"pr.provider_id",
	"pr.status",
	"pr.reviewer_id",
	"pr.comments",
	"pr.reviewed_at",
	"pr.created_at",
	"pr.updated_at",
}

func InsertProviderRequest(providerID int64) sq.InsertBuilder {
	return psql.Insert(providerRequestsTableName).
		Columns("provider_id").
		Values(providerID).
		Suffix("RETURNING id")
}

func GetProviderRequestByID(id string) sq.SelectBuilder {
	return psql.Select(providerRequestSelectCommon...).
		From(fmt.Sprintf("%s AS pr", providerRequestsTableName)).
		Where(sq.Eq{"pr.id": id}).
		Where(sq.Eq{"pr.deleted_at": nil})
}

func GetProviderRequests(faculty string, perPage, offset uint64) sq.SelectBuilder {
	q := psql.Select(providerRequestSelectCommon...).
		From(fmt.Sprintf("%s AS pr", providerRequestsTableName)).
		Join(fmt.Sprintf("%s AS p ON p.id = pr.provider_id", providersTableName)).
		Where(sq.Eq{"pr.deleted_at": nil})

	if faculty != "" {
		q = q.Where(sq.Eq{"p.faculty": faculty})
	}

	return q.OrderBy("pr.created_at DESC").
		Limit(perPage).
		Offset(offset)
}

func CountProviderRequests() sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(providerRequestsTableName).
		Where(sq.Eq{"deleted_at": nil})
}

func ApproveProviderRequest(id, reviewerID string) sq.UpdateBuilder {
	return psql.Update(providerRequestsTableName).
		Set("status", "approved").
		Set("reviewer_id", reviewerID).
		Set("reviewed_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}

func RejectProviderRequest(id, reviewerID, comments string) sq.UpdateBuilder {
	return psql.Update(providerRequestsTableName).
		Set("status", "rejected").
		Set("reviewer_id", reviewerID).
		Set("comments", comments).
		Set("reviewed_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id})
}
