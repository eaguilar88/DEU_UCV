package queries

import (
	sq "github.com/Masterminds/squirrel"
)

func GetDeuActiveGroupsCount() sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(groupsTableName + " AS g").
		Where(sq.Eq{"g.is_active": true, "g.deleted_at": nil})
}

func GetDeuInactiveGroupsCount() sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(groupsTableName + " AS g").
		Where(sq.Eq{"g.is_active": false, "g.deleted_at": nil})
}

func GetDeuPendingRequestsCount() sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(groupRequestsTableName + " AS gr").
		Where(sq.Eq{"gr.status": "under_review"})
}

func CountGroupsByFaculty(faculty string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(groupsTableName + " AS g").
		Where("? = ANY(g.faculty)", faculty).
		Where(sq.Eq{"g.deleted_at": nil})
}
