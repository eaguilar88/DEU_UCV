package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var providerQuerySelectCommon = []string{
	"p.id",
	"u.id",
	"u.first_name",
	"u.last_name",
	"p.code",
	"p.is_active",
	"p.created_at",
	"p.updated_at",
}

func GetProviderByID(id string, status, isDeleted bool) sq.SelectBuilder {
	builder := psql.Select(providerQuerySelectCommon...).
		From(fmt.Sprintf("%s AS p", providersTableName)).
		Join(fmt.Sprintf("%s AS u ON u.id = p.user_id", usersTableName)).
		Where(sq.Eq{"p.is_active": status}).
		Where(sq.Eq{"p.id": id})

	if isDeleted {
		builder = builder.Where(sq.NotEq{"p.deleted_at": nil})
	}
	return builder
}

func GetProviderByCode(code string) sq.SelectBuilder {
	return psql.Select(providerQuerySelectCommon...).
		From(fmt.Sprintf("%s AS p", providersTableName)).
		Join(fmt.Sprintf("%s AS u ON u.id = p.user_id", usersTableName)).
		Where(sq.Eq{"p.code": code})
}

func GetProviders(limit, offset int) sq.SelectBuilder {
	return psql.Select(providerQuerySelectCommon...).
		From(fmt.Sprintf("%s AS p", providersTableName)).
		Join(fmt.Sprintf("%s AS u ON u.id = p.user_id", usersTableName)).
		OrderBy("p.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func CreateProvider(provider models.Provider) sq.InsertBuilder {
	return psql.Insert(providersTableName).
		Columns(
			"user_id",
			"code",
		).
		Values(
			provider.UserID,
			provider.Code,
		).
		Suffix("RETURNING id")
}

func UpdateProvider(provider models.Provider) sq.UpdateBuilder {
	return psql.Update(providersTableName).
		Set("user_id", provider.UserID).
		Set("code", provider.Code).
		Set("is_active", provider.Active).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": provider.ID})
}

func DeleteProvider(id string) sq.UpdateBuilder {
	return psql.Update(providersTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Set("is_active", false).
		Where(sq.Eq{"id": id})
}
