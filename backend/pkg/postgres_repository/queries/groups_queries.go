package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/pkg/postgres_repository/models"
)

var (
	groupQuerySelectCommon = []string{
		"g.id",
		"g.name",
		"g.description",
		"g.endorsement_id",
		"owner.id",
		"owner.first_name",
		"owner.last_name",
		"reviewer.id",
		"reviewer.first_name",
		"reviewer.last_name",
		"g.objective",
		"g.location",
		"g.created_at",
		"g.updated_at",
		"g.deleted_at",
	}
)

func GetGroupByID(groupID string) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(fmt.Sprintf("%s AS g", groupsTableName)).
		Join(fmt.Sprintf("%s AS r ON g.endorsement_id = r.id", endorsementsTableName)).
		Join(fmt.Sprintf("%s AS owner ON g.user_id = owner.id", usersTableName)).
		Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Where(sq.Eq{"g.id": groupID})
}

func GetGroups(limit, offset int) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(groupsTableName).
		Limit(uint64(limit)).
		Offset(uint64(offset))
}

func InsertGroup(group models.ExtensionGroup) sq.InsertBuilder {
	return psql.Insert(groupsTableName).
		Columns(
			"name",
			"description",
			"user_id",
			"endorsement_id",
			"objective",
			"location",
		).
		Values(
			group.Name,
			group.Description,
			group.OwnerID,
			group.EndorsementID,
			group.Objective,
			group.Location,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func UpdateGroup(group models.ExtensionGroup) sq.UpdateBuilder {
	return psql.Update(groupsTableName).
		Set("name", group.Name).
		Set("description", group.Description).
		Set("endorsement_id", group.EndorsementID).
		Set("objective", group.Objective).
		Set("location", group.Location).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": group.ID})
}

func DeleteGroup(groupID string) sq.DeleteBuilder {
	return psql.Delete(groupsTableName).
		Where(sq.Eq{"id": groupID})
}
