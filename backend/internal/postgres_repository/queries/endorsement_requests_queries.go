package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var (
	endorsementQuerySelectCommon = []string{
		"r.id",
		"r.name",
		"r.description",
		"r.type",
		"r.status",
		"r.comments",
		"owner.id owner_id",
		"owner.first_name",
		"owner.last_name",
		"reviewer.id reviewer_id",
		"reviewer.first_name",
		"reviewer.last_name",
		"r.reviewed_at",
		"r.created_at",
		"r.updated_at",
	}
)

func GetEndorsementByID(endorsementID string) sq.SelectBuilder {
	return psql.Select(endorsementQuerySelectCommon...).
		From(fmt.Sprintf("%s AS r", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS owner ON r.user_id = owner.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Where(sq.Eq{"r.id": endorsementID})
}

func GetEndorsements(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(endorsementQuerySelectCommon...).
		From(fmt.Sprintf("%s AS r", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS owner ON r.user_id = owner.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertEndorsement(endorsement models.EndorsementRequest) sq.InsertBuilder {
	return psql.Insert(endorsementsTableName).
		Columns(
			"user_id",
			"type",
			"name",
			"description",
			"status",
		).
		Values(
			endorsement.UserID,
			endorsement.Type,
			endorsement.Name.String,
			endorsement.Description.String,
			entities.EndorsementStatus_CREATED,
		).Suffix("RETURNING id")
}

func UpdateEndorsementInfo(endorsement models.EndorsementRequest, endorsementID string) sq.UpdateBuilder {
	return psql.Update(endorsementsTableName).
		Set("user_id", endorsement.UserID).
		Set("status", endorsement.Status).
		Set("type", endorsement.Type).
		Set("name", endorsement.Name.String).
		Set("description", endorsement.Description.String).
		Set("comments", endorsement.Comments.String).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"r.id": endorsementID})
}

func DeleteEndorsement(endorsementID string) sq.DeleteBuilder {
	return psql.Delete(endorsementsTableName).
		Where(sq.Eq{"r.id": endorsementID})
}
