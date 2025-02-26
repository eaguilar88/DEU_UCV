package queries

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
)

var (
	endorsementQuerySelectCommon = []string{
		"id", "user_id", "type", "name", "description", "status", "comments", "created_at", "updated_at",
	}
)

func GetEndorsementByID(endorsementID int) sq.SelectBuilder {
	return psql.Select(endorsementQuerySelectCommon...).
		From(endorsementsTableName).
		Where(sq.Eq{"id": endorsementID})
}

func GetEndorsements(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(endorsementQuerySelectCommon...).
		From(endorsementsTableName).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertEndorsement(endorsement models.Endorsement) sq.InsertBuilder {
	return psql.Insert(endorsementsTableName).
		Columns(
			"user_id",
			"type",
			"name",
			"description",
			"status",
			"comments",
			"created_at",
			"updated_at",
		).
		Values(
			endorsement.UserID,
			endorsement.Type,
			endorsement.Name.String,
			endorsement.Description.String,
			endorsement.Status,
			endorsement.Comments.String,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func UpdateEndorsementInfo(endorsement models.Endorsement, endorsementID int) sq.UpdateBuilder {
	return psql.Update(endorsementsTableName).
		Set("user_id", endorsement.UserID).
		Set("status", endorsement.Status).
		Set("type", endorsement.Type).
		Set("name", endorsement.Name.String).
		Set("description", endorsement.Description.String).
		Set("comments", endorsement.Comments.String).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": endorsementID})
}

func DeleteEndorsement(endorsementID int) sq.DeleteBuilder {
	return psql.Delete(endorsementsTableName).
		Where(sq.Eq{"id": endorsementID})
}
