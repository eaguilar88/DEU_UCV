package queries

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
)

var activitySelectCols = []string{
	"a.id",
	"a.group_id",
	"a.name",
	"a.description",
	"a.date",
	"a.knowledge_area",
	"a.allies",
	"a.stimated_participants",
	"a.actual_participants",
	"a.financing",
	"a.comments",
	"a.created_at",
	"a.updated_at",
	"a.deleted_at",
}

func GetActivityByID(id string) sq.SelectBuilder {
	return psql.Select(activitySelectCols...).
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.id": id, "a.deleted_at": nil})
}

func GetActivities(filter entities.ActivityFilter, limit, offset int) sq.SelectBuilder {
	q := psql.Select(activitySelectCols...).
		From(activitiesTableName + " AS a").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	if filter.GroupID != "" {
		q = q.Where(sq.Eq{"a.group_id": filter.GroupID})
	}
	if filter.Date != "" {
		q = q.Where(sq.Eq{"a.date": filter.Date})
	}
	if filter.ActualParticipants != nil {
		q = q.Where(sq.Eq{"a.actual_participants": *filter.ActualParticipants})
	}
	if !filter.Deleted {
		q = q.Where(sq.Eq{"a.deleted_at": nil})
	}
	return q
}

func CountActivities(filter entities.ActivityFilter) sq.SelectBuilder {
	q := psql.Select("COUNT(*)").
		From(activitiesTableName + " AS a")

	if filter.GroupID != "" {
		q = q.Where(sq.Eq{"a.group_id": filter.GroupID})
	}
	if filter.Date != "" {
		q = q.Where(sq.Eq{"a.date": filter.Date})
	}
	if filter.ActualParticipants != nil {
		q = q.Where(sq.Eq{"a.actual_participants": *filter.ActualParticipants})
	}
	if !filter.Deleted {
		q = q.Where(sq.Eq{"a.deleted_at": nil})
	}
	return q
}

func InsertActivity(a models.Activity) sq.InsertBuilder {
	return psql.Insert(activitiesTableName).
		Columns(
			"group_id",
			"name",
			"description",
			"date",
			"knowledge_area",
			"allies",
			"stimated_participants",
			"actual_participants",
			"financing",
			"comments",
			"created_at",
			"updated_at",
		).
		Values(
			a.GroupID,
			a.Name,
			a.Description,
			a.Date,
			a.KnowledgeArea,
			a.Allies,
			a.EstimatedParticipants,
			a.ActualParticipants,
			a.Financing,
			a.Comments,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func UpdateActivity(a models.Activity) sq.UpdateBuilder {
	return psql.Update(activitiesTableName).
		Set("name", a.Name).
		Set("description", a.Description).
		Set("date", a.Date).
		Set("knowledge_area", a.KnowledgeArea).
		Set("allies", a.Allies).
		Set("stimated_participants", a.EstimatedParticipants).
		Set("actual_participants", a.ActualParticipants).
		Set("financing", a.Financing).
		Set("comments", a.Comments).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": a.ID, "deleted_at": nil})
}

func SoftDeleteActivity(id string) sq.UpdateBuilder {
	return psql.Update(activitiesTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "deleted_at": nil})
}
