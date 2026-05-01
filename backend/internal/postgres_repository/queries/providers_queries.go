package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var providerQuerySelectCommon = []string{
	"p.id",
	"p.user_id",
	"p.name",
	"p.party_type",
	"p.profit_type",
	"p.is_internal",
	"p.bio",
	"p.code",
	"p.is_active",
	"p.created_at",
	"p.updated_at",
	"p.deleted_at",
	"u.email",
	"u.first_name",
	"u.last_name",
	"p.faculty",
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

func GetProviderByUserID(userID string) sq.SelectBuilder {
	return psql.Select(providerQuerySelectCommon...).
		From(fmt.Sprintf("%s AS p", providersTableName)).
		Join(fmt.Sprintf("%s AS u ON u.id = p.user_id", usersTableName)).
		Where(sq.Eq{"p.user_id": userID})
}

func GetProviderCodeByUserID(userID string) sq.SelectBuilder {
	return psql.Select("COALESCE(p.code, '')").
		From(fmt.Sprintf("%s AS p", providersTableName)).
		Where(sq.Eq{"p.deleted_at": nil}).
		Where(sq.Eq{"p.user_id": userID})
}

func GetProviders(filters entities.ProviderFilters, limit, offset int) sq.SelectBuilder {
	q := psql.Select(providerQuerySelectCommon...).
		From(fmt.Sprintf("%s AS p", providersTableName)).
		Join(fmt.Sprintf("%s AS u ON u.id = p.user_id", usersTableName)).
		OrderBy("p.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	switch filters.Type {
	case entities.CourseProviderType:
		q = q.Where(sq.Like{"p.code": "ECP-%"})
	case entities.GroupProviderType:
		q = q.Where(sq.Like{"p.code": "GEX-%"})
	}
	if filters.PartyType != "" {
		q = q.Where(sq.Eq{"p.party_type": string(filters.PartyType)})
	}
	if filters.ProfitType != "" {
		q = q.Where(sq.Eq{"p.profit_type": string(filters.ProfitType)})
	}
	if filters.IsInternal != nil {
		q = q.Where(sq.Eq{"p.is_internal": *filters.IsInternal})
	}
	if filters.IsActive != nil {
		q = q.Where(sq.Eq{"p.is_active": *filters.IsActive})
	}
	if filters.Code != "" {
		q = q.Where(sq.Eq{"p.code": filters.Code})
	}
	if filters.CreatedAtFrom != "" {
		q = q.Where(sq.GtOrEq{"p.created_at": filters.CreatedAtFrom})
	}
	if filters.CreatedAtTo != "" {
		q = q.Where(sq.LtOrEq{"p.created_at": filters.CreatedAtTo})
	}
	return q
}

func CreateProvider(provider models.Provider) sq.InsertBuilder {
	return psql.Insert(providersTableName).
		Columns(
			"user_id",
			"name",
			"party_type",
			"profit_type",
			"is_internal",
			"bio",
			"code",
			"faculty",
		).
		Values(
			provider.UserID,
			provider.Name,
			provider.PartyType,
			provider.ProfitType,
			provider.IsInternal,
			provider.Bio,
			provider.Code,
			provider.Faculty,
		).
		Suffix("RETURNING id")
}

func UpdateProvider(provider models.Provider) sq.UpdateBuilder {
	return psql.Update(providersTableName).
		Set("user_id", provider.UserID).
		Set("name", provider.Name).
		Set("party_type", provider.PartyType).
		Set("profit_type", provider.ProfitType).
		Set("is_internal", provider.IsInternal).
		Set("bio", provider.Bio).
		Set("code", provider.Code).
		Set("is_active", provider.IsActive).
		Set("faculty", provider.Faculty).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": provider.ID})
}

func DeleteProvider(id string) sq.UpdateBuilder {
	return psql.Update(providersTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Set("is_active", false).
		Where(sq.Eq{"id": id})
}

func HardDeleteProvider(id string) sq.DeleteBuilder {
	return psql.Delete(providersTableName).
		Where(sq.Eq{"id": id})
}
