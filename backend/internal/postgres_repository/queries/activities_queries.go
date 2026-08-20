package queries

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/postgres_repository/models"
	"github.com/lib/pq"
)

var activitySelectCols = []string{
	"a.id",
	"a.group_id",
	"g.name AS group_name",
	"a.name",
	"a.description",
	"a.date_start",
	"a.date_end",
	"a.location",
	"a.knowledge_area",
	"a.allies",
	"a.group_participants",
	"a.stimated_participants",
	"a.actual_participants",
	"a.financing",
	"a.comments",
	"a.gallery_url",
	"a.report_checked",
	"a.is_featured",
	"a.created_at",
	"a.updated_at",
	"a.deleted_at",
}

func GetActivityByID(id string) sq.SelectBuilder {
	return psql.Select(activitySelectCols...).
		From(activitiesTableName + " AS a").
		LeftJoin("deu.extension_groups AS g ON a.group_id = g.id").
		Where(sq.Eq{"a.id": id, "a.deleted_at": nil})
}

func GetActivities(filter entities.ActivityFilter, limit, offset int) sq.SelectBuilder {
	q := psql.Select(activitySelectCols...).
		From(activitiesTableName + " AS a").
		LeftJoin("deu.extension_groups AS g ON a.group_id = g.id")

	q = applyActivityFilters(q, filter)

	if filter.Order == "asc" {
		q = q.OrderBy("a.date_start ASC")
	} else {
		q = q.OrderBy("a.date_start DESC")
	}

	if !filter.DisablePaging {
		q = q.Limit(uint64(limit)).Offset(uint64(offset))
	}

	return q
}

func CountActivities(filter entities.ActivityFilter) sq.SelectBuilder {
	q := psql.Select("COUNT(*)").
		From(activitiesTableName + " AS a")

	return applyActivityFilters(q, filter)
}

func GetActivityMetrics(groupID string) sq.SelectBuilder {
	q := psql.Select(
		"COUNT(*)",
		"COUNT(*) FILTER (WHERE CURRENT_DATE < a.date_start)",
		"COUNT(*) FILTER (WHERE CURRENT_DATE BETWEEN a.date_start AND a.date_end)",
		"COUNT(*) FILTER (WHERE CURRENT_DATE > a.date_end AND (a.actual_participants IS NULL OR a.actual_participants = 0))",
	).From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.deleted_at": nil})

	if groupID != "" {
		q = q.Where(sq.Eq{"a.group_id": groupID})
	}

	return q
}

func applyActivityFilters(q sq.SelectBuilder, filter entities.ActivityFilter) sq.SelectBuilder {
	if filter.GroupID != "" {
		q = q.Where(sq.Eq{"a.group_id": filter.GroupID})
	}
	if filter.NameSearch != "" {
		q = q.Where(sq.ILike{"a.name": fmt.Sprintf("%%%s%%", filter.NameSearch)})
	}

	// --- LÓGICA DE Detección Puntual vs Rango ---
	if filter.StartDate != "" && filter.EndDate != "" && filter.StartDate == filter.EndDate {
		// Caso: Consulta de un solo día (Ej: status "en_curso")
		// La fecha actual debe estar DENTRO del rango de la actividad:
		// a.date_start <= hoy AND a.date_end >= hoy
		q = q.Where(sq.LtOrEq{"a.date_start": filter.StartDate})
		q = q.Where(sq.GtOrEq{"a.date_end": filter.EndDate})
	} else {
		// Caso: Filtros por rangos de fecha distintos
		if filter.StartDate != "" {
			q = q.Where(sq.GtOrEq{"a.date_start": filter.StartDate})
		}
		if filter.EndDate != "" {
			q = q.Where(sq.LtOrEq{"a.date_end": filter.EndDate})
		}
	}

	if filter.ActualParticipants != nil {
		q = q.Where(sq.Eq{"a.actual_participants": *filter.ActualParticipants})
	}
	if filter.HasActualParticipants != nil {
		if *filter.HasActualParticipants {
			q = q.Where(sq.And{
				sq.NotEq{"a.actual_participants": nil},
				sq.Gt{"a.actual_participants": 0},
			})
		} else {
			q = q.Where(sq.Or{
				sq.Eq{"a.actual_participants": nil},
				sq.Eq{"a.actual_participants": 0},
			})
		}
	}
	if filter.IsFeatured != nil {
		q = q.Where(sq.Eq{"a.is_featured": *filter.IsFeatured})
	}
	if filter.ReportChecked != nil {
		q = q.Where(sq.Eq{"a.report_checked": *filter.ReportChecked})
	}
	if !filter.Deleted {
		q = q.Where(sq.Eq{"a.deleted_at": nil})
	}
	return q
}

// Queries para el dashboard
func CountPlannedActivitiesByGroupID(groupID string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.group_id": groupID, "a.deleted_at": nil}).
		Where("CURRENT_DATE < a.date_start")
}

func CountPendingReportActivitiesByGroupID(groupID string) sq.SelectBuilder {
	return psql.Select("COUNT(*)").
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.group_id": groupID, "a.deleted_at": nil}).
		Where("CURRENT_DATE > a.date_end").
		Where(sq.Or{
			sq.Eq{"a.actual_participants": nil},
			sq.Eq{"a.actual_participants": 0},
		})
}

func GetCurrentActivitiesByGroupID(groupID string) sq.SelectBuilder {
	return psql.Select("a.id", "a.name").
		From(activitiesTableName + " AS a").
		Where(sq.Eq{"a.group_id": groupID, "a.deleted_at": nil}).
		Where("CURRENT_DATE BETWEEN a.date_start AND a.date_end")
}

func InsertActivity(a models.Activity) sq.InsertBuilder {
	return psql.Insert(activitiesTableName).
		Columns(
			"group_id",
			"name",
			"description",
			"date_start",
			"date_end",
			"location",
			"knowledge_area",
			"allies",
			"group_participants",
			"stimated_participants",
			"actual_participants",
			"financing",
			"comments",
			"gallery_url",
			"report_checked",
			"is_featured",
			"created_at",
			"updated_at",
		).
		Values(
			a.GroupID,
			a.Name,
			a.Description,
			a.DateStart,
			a.DateEnd,
			a.Location,
			pq.Array(a.KnowledgeArea),
			a.Allies,
			a.GroupParticipants,
			a.EstimatedParticipants,
			a.ActualParticipants,
			a.Financing,
			a.Comments,
			a.GalleryURL,
			a.ReportChecked,
			a.IsFeatured,
			sq.Expr("NOW()"),
			sq.Expr("NOW()"),
		).Suffix("RETURNING id")
}

func UpdateActivity(a models.Activity) sq.UpdateBuilder {
	return psql.Update(activitiesTableName).
		Set("name", a.Name).
		Set("description", a.Description).
		Set("date_start", a.DateStart).
		Set("date_end", a.DateEnd).
		Set("location", a.Location).
		Set("knowledge_area", pq.Array(a.KnowledgeArea)).
		Set("allies", a.Allies).
		Set("group_participants", a.GroupParticipants).
		Set("stimated_participants", a.EstimatedParticipants).
		Set("actual_participants", a.ActualParticipants).
		Set("financing", a.Financing).
		Set("comments", a.Comments).
		Set("gallery_url", a.GalleryURL).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": a.ID, "deleted_at": nil})
}

func UpdateReportCheckStatus(id string, checked bool) sq.UpdateBuilder {
	return psql.Update(activitiesTableName).
		Set("report_checked", checked).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "deleted_at": nil})
}

func UpdateFeatureStatus(id string, featured bool) sq.UpdateBuilder {
	return psql.Update(activitiesTableName).
		Set("is_featured", featured).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "deleted_at": nil})
}

func SoftDeleteActivity(id string) sq.UpdateBuilder {
	return psql.Update(activitiesTableName).
		Set("deleted_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "deleted_at": nil})
}
