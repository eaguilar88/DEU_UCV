package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var groupQuerySelectCommon = []string{
	"g.id",
	"g.user_id",
	"g.name",
	"g.description",
	"g.faculty",
	"g.foundation",
	"g.is_multidisciplinary",
	"g.objective",
	"g.code",
	"g.group_director",
	"g.type",
	"g.location",
	"g.is_active",
	"g.created_at",
	"g.updated_at",
	"g.deleted_at",
}

func GetGroupByID(groupID string) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(fmt.Sprintf("%s AS g", groupsTableName)).
		//Join(fmt.Sprintf("%s AS r ON g.request_id = r.id", courseRequestsTableName)).
		Join(fmt.Sprintf("%s AS owner ON g.user_id = owner.id", usersTableName)).
		//Join(fmt.Sprintf("%s AS reviewer ON r.reviewer_id = reviewer.id", usersTableName)).
		Where(sq.Eq{"g.id": groupID})
}

func GetGroupByUserID(userID string) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(fmt.Sprintf("%s AS g", groupsTableName)).
		Join(fmt.Sprintf("%s AS owner ON g.user_id = owner.id", usersTableName)).
		Where(sq.Eq{"g.user_id": userID}).
		Where(sq.Eq{"g.deleted_at": nil}).
		OrderBy("g.created_at ASC").
		Limit(1)
}

func GetGroups(filter entities.GroupFilter, limit, offset int) sq.SelectBuilder {
	q := psql.Select(groupQuerySelectCommon...).
		From(fmt.Sprintf("%s AS g", groupsTableName)).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		OrderBy("g.name ASC")

	if filter.Faculty != "" {
		q = q.Where("? = ANY(g.faculty)", string(filter.Faculty))
	}
	if filter.Type != "" {
		q = q.Where("? = ANY(g.type)", string(filter.Type))
	}
	if filter.Active != nil {
		q = q.Where(sq.Eq{"g.is_active": *filter.Active})
	}
	if filter.Search != "" {
		// Búsqueda case-insensitive que coincida con cualquier parte del nombre
		q = q.Where(sq.ILike{"g.name": fmt.Sprintf("%%%s%%", filter.Search)})
	}
	if !filter.Deleted {
		q = q.Where(sq.Eq{"g.deleted_at": nil})
	}
	return q
}

func GetRandomActiveGroups(limit int) sq.SelectBuilder {
	return psql.Select(groupQuerySelectCommon...).
		From(fmt.Sprintf("%s AS g", groupsTableName)).
		Where(sq.Eq{"g.is_active": true}).
		OrderBy("random()").
		Limit(uint64(limit))
}

func GetGroupsSimple() sq.SelectBuilder {
	return psql.Select("g.id", "g.name").
		From(fmt.Sprintf("%s AS g", groupsTableName)).
		Where(sq.Eq{"g.is_active": true}).
		Where(sq.Eq{"g.deleted_at": nil}).
		OrderBy("g.name ASC")
}

func InsertGroup(group models.ExtensionGroup) sq.InsertBuilder {
	return psql.Insert(groupsTableName).
		Columns(
			"user_id",
			"name",
			"description",
			"faculty",
			"foundation",
			"is_multidisciplinary",
			"objective",
			"code",
			"group_director",
			"type",
			"location",
			"is_active",
			"created_at",
			"updated_at",
		).
		Values(
			group.UserID,
			group.Name,
			group.Description,
			group.Faculty,
			group.Foundation,
			group.IsMultidisciplinary,
			group.Objective,
			group.Code,
			group.Director,
			group.Type,
			group.Location,
			group.IsActive,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func UpdateGroup(group models.ExtensionGroup) sq.UpdateBuilder {
	return psql.Update(groupsTableName).
		Set("description", group.Description).
		Set("objective", group.Objective).
		Set("type", group.Type).
		Set("location", group.Location).
		Set("is_active", group.IsActive).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": group.ID})
}

func DeleteGroup(groupID string) sq.DeleteBuilder {
	return psql.Delete(groupsTableName).
		Where(sq.Eq{"id": groupID})
}

var groupMembersTableName = fmt.Sprintf("%s.group_members", schema)

func InsertGroupMember(groupID int64, m models.GroupMember) sq.InsertBuilder {
	return psql.Insert(groupMembersTableName).
		Columns("group_id", "name", "ci", "phone", "email", "coordination", "year", "faculty", "school", "is_leader", "is_active").
		Values(groupID, m.Name, m.CI, m.Phone, m.Email, m.Coordination, m.Year, m.Faculty, m.School, m.IsLeader, m.IsActive).
		Suffix("RETURNING id")
}

func UpdateGroupMember(groupID int64, m models.GroupMember) sq.UpdateBuilder {
	return psql.Update(groupMembersTableName).
		Set("name", m.Name).
		Set("phone", m.Phone).
		Set("email", m.Email).
		Set("coordination", m.Coordination).
		Set("year", m.Year).
		Set("faculty", m.Faculty).
		Set("school", m.School).
		Set("is_leader", m.IsLeader).
		Set("is_active", m.IsActive).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": m.ID, "group_id": groupID}) // group_id como cinturón de seguridad
}

func SelectGroupMembers(groupID string) sq.SelectBuilder {
	return psql.Select("id", "name", "ci", "phone", "email", "coordination", "year", "faculty", "school", "is_leader", "is_active").
		From(groupMembersTableName).
		Where(sq.Eq{"group_id": groupID, "deleted_at": nil})
}
