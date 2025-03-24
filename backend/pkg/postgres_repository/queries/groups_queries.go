package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
)

var (
	groupQuerySelectCommon = []string{
		"g.id",
		"e.name",
		"e.description",
		"e.endorsement_id",
		"requester.id",
		"requester.first_name",
		"requester.last_name",
		"reviewer.id",
		"reviewer.first_name",
		"reviewer.last_name",
		"c.objectives",
		"c.content",
		"c.cost",
		"c.location",
		"c.created_at",
	}
)

func GetGroupByID(groupID int) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(fmt.Sprintf("%s AS e", entitiesTableName)).
		Join(fmt.Sprintf("%s AS g ON e.id = g.entity_id", groupsTableName)).
		Join(fmt.Sprintf("%s AS r ON e.endorsement_id = r.id", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS requester ON r.user_id = requester.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Where(sq.Eq{"g.id": groupID})
}

func GetGroups(page entities.PageScope) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(groupsTableName).
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))
}

func InsertGroup(group models.ExtensionGroup) sq.InsertBuilder {
	return psql.Insert(groupsTableName).
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
	return psql.Delete(groupsTableName).
		Where(sq.Eq{"id": endorsementID})
}
