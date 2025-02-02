package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/repository/models"
)

var (
	groupTableName         = fmt.Sprintf("%s.extension_groups", schema)
	groupQuerySelectCommon = []string{
		"c.id", "c.user_id", "c.endorsement_id", "c.endorsed_by", "c.objectives", "c.content", "c.cost", "c.location", "c.created_at",
	}
)

func GetGroupByID(groupID int) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(groupTableName).
		Where(sq.Eq{"id": groupID})
}

func GetGroups(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(groupTableName).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertGroup(group models.ExtensionGroup) sq.InsertBuilder {
	return psql.Insert(groupTableName).
		Columns(
			"user_id",
			"endorsement_id",
			"name",
			"description",
			"objective",
			"action",
			"reach",
			"path",
			"created_at",
			"updated_at",
		).
		Values(
			group.UserID,
			group.EndorsementID,
			group.Name,
			group.Description,
			group.Objective,
			group.Action,
			group.Reach,
			group.Path,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func DeleteGroup(endorsementID int) sq.DeleteBuilder {
	return psql.Delete(groupTableName).
		Where(sq.Eq{"id": endorsementID})
}
